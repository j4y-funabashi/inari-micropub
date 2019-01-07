package eventlog

import (
	"bytes"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
)

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

type PostCreatedEvent struct {
	EventID      string          `json:"eventID"`
	EventType    string          `json:"eventType"`
	EventVersion string          `json:"eventVersion"`
	EventData    mf2.MicroFormat `json:"eventData"`
}

func (e PostCreatedEvent) Apply(list mf2.PostList) mf2.PostList {
	list.Add(e.EventData)
	return list
}

type nullEvent struct{}

func (e nullEvent) Apply(list mf2.PostList) mf2.PostList {
	return list
}

// NewMutator will take an event json string and return the appropriate
// mutate function based on the eventType
func NewMutator(eventJSON string) Mutator {

	e := struct {
		EventType string `json:"eventType"`
	}{}
	err := json.Unmarshal([]byte(eventJSON), &e)
	if err != nil {
		return nullEvent{}
	}

	switch e.EventType {
	case "PostCreated":
		ev := PostCreatedEvent{}
		json.Unmarshal([]byte(eventJSON), &ev)
		return ev
	}

	return nullEvent{}
}

func NewS3EventList(bucket, prefix string, client EventListReader, logger *log.Entry) (EventList, error) {
	return func() ([]Mutator, error) {
		out := []Mutator{}
		allKeys, err := client.ListKeys(bucket, prefix)
		if err != nil {
			logger.WithError(err).Error("failed to list keys")
			return []Mutator{}, nil
		}
		logger.Infof("found %d keys", len(allKeys))
		for _, key := range allKeys {
			buf, err := client.ReadObject(*key, bucket)
			if err != nil {
				logger.WithError(err).WithField("s3Key", key).Error("failed to read file")
				continue
			}
			out = append(out, NewMutator(buf.String()))
		}
		return out, nil
	}, nil
}

// Replay builds up a PostList by fetching all events and applying them
func Replay(eventListing EventList, logger *log.Entry) (mf2.PostList, error) {
	postList := mf2.PostList{}
	events, err := eventListing()
	if err != nil {
		return postList, err
	}
	for _, event := range events {
		postList = event.Apply(postList)
		if err != nil {
			return postList, err
		}
	}
	postList.Sort()
	return postList, nil
}
