package main

import (
	"os"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/sirupsen/logrus"
)

func main() {

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	S3KeyPrefix := os.Getenv("S3_EVENTS_KEY")
	S3Bucket := os.Getenv("S3_EVENTS_BUCKET")

	// deps
	logger := logrus.New()

	s3Client, err := s3.NewClient(s3Endpoint)
	if err != nil {
		logger.WithError(err).Error("failed to connect to s3")
		return
	}

	err = s3Client.CreateBucket(S3Bucket)
	if err != nil {
		logger.WithError(err).Error("failed to create s3 bucket")
		return
	}

	sqlDB, err := db.OpenDB()
	if err != nil {
		logger.WithError(err).Error("failed to open DB")
		return
	}

	eventLog := eventlog.NewEventLog(
		S3KeyPrefix,
		S3Bucket,
		s3Client,
		sqlDB,
		logger,
	)

	body := `{"type": ["h-entry"], "properties": {"uid": ["test123"], "url": ["http://localhost:8091/test123"], "published": ["2019-01-28T13:13:13+00:00"], "photo": ["https://media.funabashi.co.uk/2019/2ae61a09e30ac24d32579471d0105c92.jpg"], "content": ["hello this is a test"]}}`
	addPost(body, eventLog)

	body = `{"type": ["h-entry"], "properties": {"uid": ["test124"], "url": ["http://localhost:8091/test124"], "published": ["2019-01-28T13:13:13+00:00"], "photo": ["https://media.funabashi.co.uk/2019/3e715dddf64a622115ca58e61c2121b1.jpg"], "content": ["hello this is a test 2"], "category": ["tag1", "tag2"]}}`
	addPost(body, eventLog)

	addMedia(eventLog)
}

func addPost(body string, eventLog eventlog.EventLog) error {
	mf, err := mf2.MfFromJson(body)
	if err != nil {
		return err
	}
	event := eventlog.NewPostCreated(mf)

	// add event to eventlog
	err = eventLog.Append(event)
	return err
}

func addMedia(eventLog eventlog.EventLog) error {
	now, _ := time.Parse("2006-Jan-02", "2018-Jan-28")
	mediaMeta := mf2.MediaMetadata{
		FileKey:  "2019/b8ea8e3ce769f2a54454d3818f90bbbf.jpg",
		MimeType: "image/jpeg",
		Uid:      "12345",
		DateTime: &now,
		Lat:      53.800755,
		Lng:      -1.549077,
	}
	event := eventlog.NewMediaUploaded(mediaMeta)
	err := eventLog.Append(event)
	return err
}
