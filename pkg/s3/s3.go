package s3

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

type Client struct {
	s3Client   *s3.S3
	bucket     string
	downloader *s3manager.Downloader
	uploader   *s3manager.Uploader
}

func NewClient() (Client, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(endpoints.EuCentral1RegionID),
	})
	downloader := s3manager.NewDownloader(s)
	uploader := s3manager.NewUploader(s)
	s3Client := s3.New(s)
	return Client{
		s3Client:   s3Client,
		downloader: downloader,
		uploader:   uploader,
	}, nil
}

func (client Client) ListKeys(bucket, prefix string) ([]*string, error) {
	var l []*string

	i := 0
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	err := client.s3Client.ListObjectsV2Pages(
		input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, item := range page.Contents {
				l = append(l, item.Key)
				i++
			}
			if lastPage == true {
				return false
			}
			return true
		},
	)
	if err != nil {
		return l, err
	}
	return l, nil
}

func (client Client) ReadObject(key, bucket string) (*bytes.Buffer, error) {
	// read event data
	in := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := client.downloader.Download(buf, &in)
	if err != nil {
		return &bytes.Buffer{}, err
	}
	return bytes.NewBuffer(buf.Bytes()), nil
}

func (client Client) WriteObject(key, bucket string, body io.Reader) error {
	_, err := client.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
		ACL:    aws.String("private"),
	})
	if err != nil {
		log.Printf("failed to upload to s3 %v", err)
		return err
	}
	return nil
}
