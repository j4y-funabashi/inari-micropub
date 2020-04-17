package s3

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Client struct {
	s3Client   *s3.S3
	bucket     string
	downloader *s3manager.Downloader
	uploader   *s3manager.Uploader
}

func NewClient(s3Endpoint string) (Client, error) {

	region := endpoints.EuCentral1RegionID

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if s3Endpoint != "" {
		sess, _ = session.NewSession(&aws.Config{
			Credentials:      credentials.NewStaticCredentials("foo", "var", ""),
			Region:           aws.String(region),
			Endpoint:         aws.String(s3Endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		})
	}

	downloader := s3manager.NewDownloader(sess)
	uploader := s3manager.NewUploader(sess)
	s3Client := s3.New(sess)

	cl := Client{
		s3Client:   s3Client,
		downloader: downloader,
		uploader:   uploader,
	}

	if s3Endpoint != "" {
		cl.CreateBucket(os.Getenv("S3_MEDIA_BUCKET"))
		cl.CreateBucket(os.Getenv("S3_EVENTS_BUCKET"))
	}
	return cl, nil
}

func (client Client) CreateBucket(bucket string) error {
	region := endpoints.EuCentral1RegionID
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		},
	}
	_, err := client.s3Client.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return nil
			default:
				return err
			}
		}
	}

	return err
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

func (client Client) WriteObject(key, bucket string, body io.Reader, isPrivate bool) error {
	acl := "private"
	if !isPrivate {
		acl = "public-read"
	}
	_, err := client.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
		ACL:    aws.String(acl),
	})
	if err != nil {
		log.Printf("failed to upload to s3 %v", err)
		return err
	}
	return nil
}
