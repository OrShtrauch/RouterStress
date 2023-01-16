package s3

import (
	// "RouterStress/consts"
	// "context"
	// "fmt"
	// "bytes"
	// "os"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func Upload(filesName []string) error {
	//_, err := session.NewSession()
	var err error

    if err!= nil {
        return err
    }    

    //uploader := s3manager.NewUploader(sess)

	//workingDir, err := os.Getwd()

	if err != nil {
		return err
	}

	//localPath := fmt.Sprintf("%v/%v/%v", workingDir, consts.RESULTS_DIR, consts.TEST_ID)

	// for _, file := range filesName {
	// 	key := fmt.Sprintf("%v/%v", consts.TEST_ID, file)
	// 	filePath := fmt.Sprintf("%v/%v", localPath, file)

	// 	input, err := uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
	// 		Bucket: aws.String(consts.BUCKET),
	// 		Key:    aws.String(key),
	// 		Body:   filePath,
	// 	})	
	// }
	return err
}
