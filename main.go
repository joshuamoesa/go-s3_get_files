package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

//MyEvent struct
type MyEvent struct {
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
	Delimiter string `json:"delimiter"`
}

//MyResponse struct
type MyResponse struct {
	Message string `json:"result"`
}

func main() {
	lambda.Start(HandleLambdaEvent)
}

//HandleLambdaEvent function
func HandleLambdaEvent(event MyEvent) (MyResponse, error) {

	Bucketchecker(event.Bucket, event.Prefix, event.Delimiter)

	return MyResponse{Message: fmt.Sprintf("%s with prefix %s is read", event.Bucket, event.Prefix)}, nil
}

// Bucketchecker function
func Bucketchecker(bucket string, prefix string, delimiter string) {

	os.Setenv("BUCKET", "messagestore")
	os.Setenv("PATH_PREFIX", "")
	os.Setenv("TIMEZONE", "Europe/Rome")
	os.Setenv("AWS_ACCESS_KEY_ID", "")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "")
	os.Setenv("AWS_REGION", "eu-west-1")

	// utc life
	loc, err := time.LoadLocation(os.Getenv("TIMEZONE"))
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})

	// Create S3 service client
	svc := s3.New(sess)

	// Get the list of items
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String(prefix), Delimiter: aws.String(delimiter)})
	if err != nil {
		exitErrorf("[ERROR] Unable to list items in bucket %q, %v", bucket, err)
	}

	i := 0
	// Loop through the list
	for _, item := range resp.Contents {
		if *item.Size > int64(0) {
			//fmt.Print(*item)
			keyArray := strings.Split(*item.Key, "/")
			fmt.Print("File info: ")
			fmt.Print("Name: ", keyArray[len(keyArray)-1], " ")
			fmt.Print("modified at: ", (*item.LastModified).In(loc), " ")
			fmt.Print("Size: ", *item.Size/1024, "KB ")
			i++
		}
	}

	fmt.Print("[INFO] File count: ", i, " ")
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
