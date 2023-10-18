package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// PutObject - Upload object to s3 bucket
func PutObject(key, bucket, s3Class string) error {
	session := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(session)

	file, err := os.Open(key)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         file,
		StorageClass: aws.String(s3Class),
	})
	if err == nil {
		log.Print("Cache saved successfully")
	}

	return err
}

// GetObject - Get object from s3 bucket
func GetObject(key, bucket string) error {
	session := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(session)

	file, err := os.Create(key)
	if err != nil {
		return err
	}

	size, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	log.Printf("Cache downloaded successfully, containing %d bytes", size)

	return err
}

// DeleteObject - Delete object from s3 bucket
func DeleteObject(key, bucket string) error {
	session := session.Must(session.NewSession())
	service := s3.New(session)

	_, err := service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		log.Print("Cache purged successfully")
	}

	return err
}

// ObjectExists - Verify if object exists in s3
func ObjectExists(key, bucket string) (bool, error) {
	session := session.Must(session.NewSession())
	service := s3.New(session)

	if _, err := service.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		if aerr := err.(awserr.Error); aerr.Code() == ErrCodeNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func PutTag(objectKey, bucket, tagKey, tagValue string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	session := s3.NewFromConfig(cfg)

	putTaggingInput := &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Tagging: &s3_types.Tagging{
			TagSet: []s3_types.Tag{
				{
					Key:   aws.String(tagKey),
					Value: aws.String(tagValue),
				},
			},
		},
	}

	_, err = session.PutObjectTagging(context.TODO(), putTaggingInput)

	if err == nil {
		log.Printf("Object %s tag %s updated to %s", objectKey, tagKey, tagValue)
	}

	return err
}

func GetTag(objectKey, bucket, tagKey string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	session := s3.NewFromConfig(cfg)

	getTaggingInput := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	}

	tags, err := session.GetObjectTagging(context.TODO(), getTaggingInput)
	if err != nil {
		return "", err
	}

	for _, tag := range tags.TagSet {
		if *tag.Key == tagKey {
			return *tag.Value, nil
		}
	}

	return "", fmt.Errorf("Tag %s not found.", tagKey)
}
