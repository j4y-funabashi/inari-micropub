package eventlog

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"path"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
)

func NewEventLog(
	s3KeyPrefix string,
	s3Bucket string,
	s3Client s3.Client,
	db *sql.DB,
	logger *logrus.Logger,
) EventLog {
	return EventLog{
		s3KeyPrefix: s3KeyPrefix,
		s3Bucket:    s3Bucket,
		s3Client:    s3Client,
		db:          db,
		logger:      logger,
	}
}

type Event interface {
	getJSON() io.Reader
	getFilekey() string
	save(sqlClient *sql.Tx, s3Client s3.Client, s3KeyPrefix, s3Bucket string) error
	reduce(sqlClient *sql.Tx) error
}

type EventLog struct {
	s3Bucket    string
	s3KeyPrefix string
	s3Client    s3.Client
	db          *sql.DB
	logger      *logrus.Logger
}

type EventListReader interface {
	ListKeys(bucket, prefix string) ([]*string, error)
	ReadObject(key, bucket string) (*bytes.Buffer, error)
}

// Mutator Applies a change to a PostList
type Mutator interface {
	Apply(list mf2.PostList) mf2.PostList
}

// EventList Returns a list of events
type EventList func() ([]Mutator, error)

type MediaUploadedEvent struct {
	EventID      string            `json:"eventID"`
	EventType    string            `json:"eventType"`
	EventVersion string            `json:"eventVersion"`
	EventData    mf2.MediaMetadata `json:"eventData"`
}

func (e MediaUploadedEvent) getJSON() io.Reader {

	eventjson := new(bytes.Buffer)
	err := json.NewEncoder(eventjson).Encode(e)
	if err != nil {
		return new(bytes.Buffer)
	}

	return eventjson
}

func (e MediaUploadedEvent) getFilekey() string {
	return e.EventVersion + "_" + e.EventID + ".json"
}

func (e MediaUploadedEvent) save(sqlClient *sql.Tx, s3Client s3.Client, s3KeyPrefix, s3Bucket string) error {
	eventData := e.getJSON()
	fileKey := path.Join(
		s3KeyPrefix,
		time.Now().Format("2006"),
		e.getFilekey(),
	)

	err := s3Client.WriteObject(
		fileKey,
		s3Bucket,
		eventData,
		true,
	)

	return err
}

func (e MediaUploadedEvent) reduce(sqlClient *sql.Tx) error {

	// TODO get this from ENV
	e.EventData.URL = "https://media.funabashi.co.uk/" + e.EventData.FileKey

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(e.EventData)
	if err != nil {
		return err
	}

	_, err = sqlClient.Exec(
		`INSERT INTO media
			(id, year, month, day, data, sort_key)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`,
		e.EventData.URL,
		e.EventData.DateTime.Format("2006"),
		e.EventData.DateTime.Format("01"),
		e.EventData.DateTime.Format("02"),
		buf.String(),
		e.EventData.DateTime.Format(time.RFC3339)+e.EventData.Uid,
	)
	return err
}

func NewMediaUploaded(data mf2.MediaMetadata) MediaUploadedEvent {
	uid := uuid.NewV4()
	return MediaUploadedEvent{
		EventID:      uid.String(),
		EventType:    "MediaUploaded",
		EventVersion: time.Now().Format("20060102150405.0000"),
		EventData:    data}
}

type MediaDeletedEvent struct {
	EventID      string `json:"eventID"`
	EventType    string `json:"eventType"`
	EventVersion string `json:"eventVersion"`
	EventData    string `json:"eventData"`
}

func NewMediaDeleted(mediaURL string) MediaDeletedEvent {
	uid := uuid.NewV4()
	return MediaDeletedEvent{
		EventID:      uid.String(),
		EventType:    "MediaDeleted",
		EventVersion: time.Now().Format("20060102150405.0000"),
		EventData:    mediaURL}
}

func (e MediaDeletedEvent) getFilekey() string {
	return e.EventVersion + "_" + e.EventID + ".json"
}

func (e MediaDeletedEvent) getJSON() io.Reader {

	eventjson := new(bytes.Buffer)
	err := json.NewEncoder(eventjson).Encode(e)
	if err != nil {
		return new(bytes.Buffer)
	}

	return eventjson
}

func (e MediaDeletedEvent) reduce(sqlClient *sql.Tx) error {
	_, err := sqlClient.Exec(
		`DELETE FROM media WHERE id = $1`,
		e.EventData,
	)
	return err
}

func (e MediaDeletedEvent) save(sqlClient *sql.Tx, s3Client s3.Client, s3KeyPrefix, s3Bucket string) error {
	eventData := e.getJSON()
	fileKey := path.Join(
		s3KeyPrefix,
		time.Now().Format("2006"),
		e.getFilekey(),
	)

	err := s3Client.WriteObject(
		fileKey,
		s3Bucket,
		eventData,
		true,
	)

	return err
}

type PostCreatedEvent struct {
	EventID      string          `json:"eventID"`
	EventType    string          `json:"eventType"`
	EventVersion string          `json:"eventVersion"`
	EventData    mf2.MicroFormat `json:"eventData"`
}

func NewPostCreated(mf mf2.MicroFormat) PostCreatedEvent {
	uid := uuid.NewV4()
	return PostCreatedEvent{
		EventID:      uid.String(),
		EventType:    "PostCreated",
		EventVersion: time.Now().Format("20060102150405.0000"),
		EventData:    mf}
}

func (e PostCreatedEvent) save(sqlClient *sql.Tx, s3Client s3.Client, s3KeyPrefix, s3Bucket string) error {
	eventData := e.getJSON()
	fileKey := path.Join(
		s3KeyPrefix,
		time.Now().Format("2006"),
		e.getFilekey(),
	)

	err := s3Client.WriteObject(
		fileKey,
		s3Bucket,
		eventData,
		true,
	)

	return err
}

func (e PostCreatedEvent) reduce(sqlClient *sql.Tx) error {

	if e.EventData.GetFirstString("photo") == "" {
		return nil
	}

	published, err := time.Parse(time.RFC3339, e.EventData.ToView().Published)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(e.EventData)
	if err != nil {
		return err
	}

	_, err = sqlClient.Exec(
		`INSERT INTO posts (id, year, month, data, sort_key) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
		e.EventData.GetFirstString("url"),
		published.Format("2006"),
		published.Format("01"),
		buf.String(),
		published.Format(time.RFC3339)+e.EventData.GetFirstString("uid"),
	)
	if err != nil {
		return err
	}

	for _, photoURL := range e.EventData.GetStringSlice("photo") {
		_, err = sqlClient.Exec(
			`INSERT INTO media_published (id) VALUES ($1) ON CONFLICT DO NOTHING`,
			photoURL,
		)
		if err != nil {
			return err
		}
	}

	return err
}

func (e PostCreatedEvent) getJSON() io.Reader {

	eventjson := new(bytes.Buffer)
	err := json.NewEncoder(eventjson).Encode(e)
	if err != nil {
		return new(bytes.Buffer)
	}

	return eventjson
}

func (e PostCreatedEvent) getFilekey() string {
	return e.EventVersion + "_" + e.EventID + ".json"
}

func (e PostCreatedEvent) Apply(list mf2.PostList) mf2.PostList {
	list.Add(e.EventData)
	return list
}

type nullEvent struct{}

func (e nullEvent) getJSON() io.Reader {
	return new(bytes.Buffer)
}
func (e nullEvent) getFilekey() string {
	return ""
}
func (e nullEvent) save(sqlClient *sql.Tx, s3Client s3.Client, s3KeyPrefix, s3Bucket string) error {
	return nil
}
func (e nullEvent) reduce(sqlClient *sql.Tx) error {
	return nil
}

// decodeEvent will take an event json string and return the appropriate
// mutate function based on the eventType
func decodeEvent(eventJSON string) (Event, error) {

	e := struct {
		EventType string `json:"eventType"`
	}{}
	err := json.Unmarshal([]byte(eventJSON), &e)
	if err != nil {
		return nullEvent{}, err
	}

	switch e.EventType {
	case "PostCreated":
		ev := PostCreatedEvent{}
		json.Unmarshal([]byte(eventJSON), &ev)
		return ev, nil
	case "MediaUploaded":
		ev := MediaUploadedEvent{}
		json.Unmarshal([]byte(eventJSON), &ev)
		return ev, nil
	}

	return nullEvent{}, nil
}

// Replay builds up a PostList by fetching all events and applying them
func (el EventLog) Replay() error {
	allKeys, err := el.s3Client.ListKeys(el.s3Bucket, el.s3KeyPrefix)
	if err != nil {
		el.logger.WithError(err).Error("failed to list event keys")
		return err
	}

	el.logger.Infof("found %d events, starting transaction", len(allKeys))
	tx, err := el.db.Begin()
	if err != nil {
		el.logger.WithError(err).Error("failed to start transaction")
		return err
	}

	for _, key := range allKeys {
		buf, err := el.s3Client.ReadObject(*key, el.s3Bucket)
		if err != nil {
			el.logger.
				WithField("key", key).
				WithError(err).
				Error("failed to read event file")
			continue
		}

		event, err := decodeEvent(buf.String())
		if err != nil {
			el.logger.
				WithField("key", key).
				WithError(err).
				Error("failed to decode event file")
			continue
		}

		err = event.reduce(tx)
		if err != nil {
			el.logger.
				WithField("key", key).
				WithError(err).
				Error("failed to reduce event to db")
			continue
		}

	}

	err = tx.Commit()
	if err != nil {
		el.logger.WithError(err).Error("failed to commit transaction")
		return err
	}
	el.logger.Infof("completed replaying %d events", len(allKeys))

	return nil
}

// Append will save an event to remote and local storage
func (el EventLog) Append(event Event) error {

	tx, err := el.db.Begin()
	if err != nil {
		el.logger.WithError(err).Error("failed to start transaction")
		return err
	}

	err = event.save(tx, el.s3Client, el.s3KeyPrefix, el.s3Bucket)
	if err != nil {
		el.logger.WithError(err).Error("failed to save event")
		return err
	}

	err = event.reduce(tx)
	if err != nil {
		el.logger.WithError(err).Error("failed to reduce event")
		return err
	}

	err = tx.Commit()
	if err != nil {
		el.logger.WithError(err).Error("failed to commit transaction")
		return err
	}

	return err
}
