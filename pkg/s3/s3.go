package s3

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/sirupsen/logrus"
)

func NewSaver(logger *logrus.Logger, s3Endpoint, S3KeyPrefix, S3Bucket string) func(event mf2.PostCreatedEvent) error {

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(endpoints.EuCentral1RegionID),
	})
	if s3Endpoint != "" {
		sess, _ = session.NewSession(&aws.Config{
			Credentials:      credentials.NewStaticCredentials("foo", "var", ""),
			Region:           aws.String(endpoints.EuCentral1RegionID),
			Endpoint:         aws.String(s3Endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		})
	}

	// deps
	uploader := s3manager.NewUploader(sess)

	return func(event mf2.PostCreatedEvent) error {

		eventjson := new(bytes.Buffer)
		err := json.NewEncoder(eventjson).Encode(event)
		if err != nil {
			logger.
				WithError(err).
				WithField("event", event).
				Error("failed to encode event to json")
			return err
		}
		eventData := eventjson.String()

		fileKey := strings.Trim(S3KeyPrefix, "/ ") + "/" + time.Now().Format("2006/") + event.EventVersion + "_" + event.EventID + ".json"

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(S3Bucket),
			Key:    aws.String(fileKey),
			Body:   eventjson,
			ACL:    aws.String("private"),
		})
		if err != nil {
			logger.
				WithError(err).
				WithField("key", fileKey).
				WithField("event", eventData).
				WithField("bucket", S3Bucket).
				Error("failed to upload event to s3")
			return err
		}

		logger.
			WithField("key", fileKey).
			WithField("event", eventData).
			WithField("bucket", S3Bucket).
			Info("uploaded event to s3")

		return nil
	}
}
