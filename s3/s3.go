package s3

import (
	"RouterStress/consts"
	"bytes"
	"fmt"	
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Upload() error {

	creds := credentials.NewEnvCredentials()

	_, err := creds.Get()
	if err != nil {
		return err
	}

	cfg := aws.NewConfig().WithRegion("eu-central-1").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	workingDir, err := os.Getwd()

	if err != nil {
		return err
	}

	testID := "Ferb_LG_Commscope_Puma7_0ddb88a5-5b40-4993-8925-c58c5832a400"
	localPath := fmt.Sprintf("%v/%v/%v", workingDir, consts.RESULTS_DIR, testID)

	files, err := os.ReadDir(localPath)

	if err != nil {
		return err
	}

	for _, file := range files {
		fileName := fmt.Sprintf("%v/%v", localPath, file)

		file, err := os.Open(fileName)

		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()

		if err != nil {
			return err
		}

		buffer := make([]byte, info.Size())

		file.Read(buffer)
		fileBytes := bytes.NewReader(buffer)
		fileType := http.DetectContentType(buffer)

		params := &s3.PutObjectInput{
			Bucket: aws.String(consts.BUCKET),
			Key: aws.String(fileName),
			Body: fileBytes,
			ContentLength: aws.Int64(info.Size()),
			ContentType: aws.String(fileType),
		}

		_, err = svc.PutObject(params)

		if err != nil {
			return err
		}
	}

	return err
}
