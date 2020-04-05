// List content of a AWS S3 bucket or prefix,
// filter out the oldest S3 object based on the LastModified attribute and
// send the age in seconds of the object as a metric to Datadog.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// MyEvent represents a JSON inputmessage containing parameters.
type MyEvent struct {
	Bucket      string `json:"bucket"`      // name of the bucket
	Prefix      string `json:"prefix"`      // optional subfolder name within the bucket
	Delimiter   string `json:"delimiter"`   // delimiter character, usually it's "/"
	Environment string `json:"environment"` // environment value like pnlt, pnla or pnlp
	MaxAgeSec   string `json:"max_age_sec"` // optional max age in seconds
}

// MyResponse represent the result coming back from the function.
type MyResponse struct {
	Message string `json:"result"`
}

func main() {
	//lambda.Start(HandleLambdaEvent)

	os.Setenv("BUCKET", "messagestore")
	os.Setenv("PATH_PREFIX", "subfolder")
	os.Setenv("DELIMITER", "")
	os.Setenv("ENVIRONMENT", "pnlt")

	bucketName := os.Getenv("BUCKET")
	pathPrefixName := os.Getenv("PATH_PREFIX")
	delimiterValue := os.Getenv("DELIMITER")
	functionnameLambdaMetricParameter := "listS3Objects"
	bucketnameMetricParameter := bucketName
	environment := os.Getenv("ENVIRONMENT")

	resp := ListBucketObjects(bucketName, pathPrefixName, delimiterValue)
	fmt.Print("[main:INFO] AWS S3 response:\n" + resp.String() + "\n")

	// assume first file lastmodified value is the smallest so it's the oldest file in seconds
	min := resp.Contents[0].LastModified.UTC().Unix()
	fmt.Print("[main:INFO] Initial oldest file: " + strconv.FormatInt(min, 10) + " \n")

	if len(pathPrefixName) > 0 && len(resp.Contents) > 1 { //when a prefix has to be taken into account, take element 1 in stead of 0
		min = resp.Contents[1].LastModified.UTC().Unix()
		fmt.Print("[main:INFO] Oldest file in a prefix: " + strconv.FormatInt(min, 10) + " \n")
		bucketnameMetricParameter = bucketnameMetricParameter + "_" + pathPrefixName
	}

	//lengthContents := strconv.Itoa(len(resp.Contents))
	//fmt.Print("Total keys: " + lengthContents + "\n")

	// Loop through the list to check if there are older files than the value in variable "min"
	for i, item := range resp.Contents {
		if len(pathPrefixName) > 0 && i == 0 { // Skip first element when dealing with prefix, assuming that the prefix is the first element
			//ignore
		} else {
			lastModified := item.LastModified.UTC().Unix()
			if lastModified < min {
				//keyArray := strings.Split(*item.Key, "/")
				//fmt.Print("Older file detected. Name: ", keyArray[len(keyArray)-1], " \n")
				fmt.Print("[main:INFO] Older object detected. Value in posix timestamp (sec) format: " + strconv.FormatInt(lastModified, 10) + " \n")
				min = lastModified
			}
		}
	}

	CreateMetric(min, bucketnameMetricParameter, functionnameLambdaMetricParameter, environment)

}

// HandleLambdaEvent is the Lambda handler signature and includes the code which will be executed.
func HandleLambdaEvent(event MyEvent) (MyResponse, error) {

	bucketName := event.Bucket
	pathPrefixName := event.Prefix
	delimiterValue := event.Delimiter
	functionnameLambdaMetricParameter := lambdacontext.FunctionName
	bucketnameMetricParameter := bucketName
	environment := event.Environment

	resp := ListBucketObjects(bucketName, pathPrefixName, delimiterValue)
	fmt.Print("[main:INFO] AWS S3 response:" + resp.String() + "\n")

	// assume first file lastmodified value is the smallest so it's the oldest file in seconds
	min := resp.Contents[0].LastModified.UTC().Unix()
	fmt.Print("[main:INFO] Initial oldest file: " + strconv.FormatInt(min, 10) + " \n")

	if len(pathPrefixName) > 0 { //when a prefix has to be taken into account, take element 1 in stead of 0
		min = resp.Contents[1].LastModified.UTC().Unix()
		fmt.Print("[main:INFO] Oldest file in a prefix: " + strconv.FormatInt(min, 10) + " \n")
		bucketnameMetricParameter = bucketnameMetricParameter + "_" + pathPrefixName
	}

	/*
	   lengthContents := strconv.Itoa(len(resp.Contents))
	   //fmt.Print("Total keys: " + lengthContents + "\n")
	*/

	// Loop through the list to check if there are older files than the value in variable "min"
	for i, item := range resp.Contents {
		if len(pathPrefixName) > 0 && i == 0 { // Skip first element when dealing with prefix, assuming that the prefix is the first element
			//ignore
		} else {
			lastModified := item.LastModified.UTC().Unix()
			if lastModified < min {
				//keyArray := strings.Split(*item.Key, "/")
				//fmt.Print("Older file detected. Name: ", keyArray[len(keyArray)-1], " \n")
				fmt.Print("[main:INFO] Older object detected. Value in posix timestamp (sec) format: " + strconv.FormatInt(lastModified, 10) + " \n")
				min = lastModified
			}
		}
	}

	createMetricResult := CreateMetric(min, bucketnameMetricParameter, functionnameLambdaMetricParameter, environment)

	return MyResponse{Message: fmt.Sprintf("S3 bucket %s with prefix %s is read. Result: %s", bucketName, pathPrefixName, createMetricResult)}, nil

}

// ListBucketObjects lists the AWS S3 bucket and/or prefix and returns a list of keys.
func ListBucketObjects(bucket string, prefix string, delimiter string) *s3.ListObjectsV2Output {

	awsAccessKey := os.Getenv("ESB_AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("ESB_AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("ESB_AWS_REGION")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, ""),
	})

	// Create S3 service client
	svc := s3.New(sess)

	// Get the list of items
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String(prefix), Delimiter: aws.String(delimiter)})
	if err != nil {
		exitErrorf("[ListBucketObjects:ERROR] Unable to list items in bucket %q, %v", bucket, err)
	}

	return resp

}

// CreateMetric creates JSON formatted string which will be sent to Datadog as a metric.
func CreateMetric(objectageSec int64, bucketName string, functionName string, environment string) string {

	url := "https://api.datadoghq." + os.Getenv("DD_SITE") + "/api/v1/series?api_key=" + os.Getenv("DD_API_KEY")
	/* Postman mockservice url
	   url := "https://5d3ee046-b510-4184-9141-4fcf2011be95.mock.pstmn.io"
	*/

	timestamp := time.Now().Unix()
	metricTimestamp := strconv.FormatInt(timestamp, 10)          // s == "97" (decimal)
	metricValue := strconv.FormatInt(timestamp-objectageSec, 10) // s == "97" (decimal)

	var jsonStr = []byte(`{"series":[{"metric":"esb.aws.s3.object.age.seconds","points":[[` +
		metricTimestamp +
		`,` +
		metricValue +
		`]],"type":"gauge","host":"` +
		bucketName +
		`","tags":["service:esb","function:` +
		functionName +
		`","environment:` +
		environment +
		`"]}]}`)

	// Build the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("DD-API-KEY", os.Getenv("DD_API_KEY"))
	req.Header.Add("DD-APPLICATION-KEY", os.Getenv("DD_APPLICATION_KEY"))

	if err != nil {
		log.Fatal("[CreateMetric:FATAL] NewRequest: ", err)
		os.Exit(1)
	}

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[CreateMetric:FATAL] Do: ", err)
		os.Exit(1)
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	fmt.Println("[CreateMetric:INFO] Metric value sent:", metricValue)
	fmt.Println("[CreateMetric:INFO] Datadog Response Status:", resp.Status)
	fmt.Println("[CreateMetric:INFO] Datadog Response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("[CreateMetric:INFO] Datadog Response Body:", string(body))

	return "OK"
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
