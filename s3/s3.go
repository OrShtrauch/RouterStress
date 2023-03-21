package s3

import (
	"RouterStress/consts"
	"fmt"	
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)


func Upload() error {
	var err error
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	})

	if err != nil {
		return err
	}

	s3Client := s3.New(session)

	workingDir, err := os.Getwd()

	if err != nil {
		return err
	}


	localPath := fmt.Sprintf("%v/%v/%v", workingDir, consts.RESULTS_DIR, consts.TEST_ID)

	files, err := os.ReadDir(localPath)

	if err != nil {
		return err
	}

	for _, file := range files {
		name := fmt.Sprintf("%v/%v", localPath, file.Name())

		fd, err := os.Open(name)

		if err != nil {
			return err
		}

		defer fd.Close()

		s3key := fmt.Sprintf("%v/%s", consts.TEST_ID, file.Name())

		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(consts.BUCKET),
			Key:    aws.String(s3key),
			Body:   fd,
		})

		if err != nil {
			return err
		}
	}

	return err
}
