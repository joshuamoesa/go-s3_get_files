package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	//lambda.Start(HandleLambdaEvent)
    //sendMetric()
    //BucketChecker
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
			fmt.Print("modified at: ", (*item.In(loc), " ")
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

func sendMetric() string {

	//DataDog Joshua
	os.Setenv("DD_API_KEY", "")
	os.Setenv("DD_APPLICATION_KEY", "")
	os.Setenv("DD_SITE", "eu")

	//DataDog SBP
	//os.Setenv("DD_API_KEY", "")
	//os.Setenv("DD_APPLICATION_KEY", "")
	//os.Setenv("DD_SITE", "com")

	url := "https://api.datadoghq." + os.Getenv("DD_SITE") + "/api/v1/series?api_key=" + os.Getenv("DD_API_KEY")
	//	url := "https://5d3ee046-b510-4184-9141-4fcf2011be95.mock.pstmn.io"

	timestampMetrics := strconv.FormatInt(time.Now().Unix(), 10) // s == "97" (decimal)

	var jsonStr = []byte(`{"series":[{"metric":"test.metric","points":[[` + timestampMetrics + `,10]],"type":"count","interval":1,"host":"test.example.com","tags":["environment:test"]}]}`)

	// Build the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("DD-API-KEY", os.Getenv("DD_API_KEY"))
	req.Header.Add("DD-APPLICATION-KEY", os.Getenv("DD_APPLICATION_KEY"))

	if err != nil {
		log.Fatal("NewRequest: ", err)
		//return
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
		log.Fatal("Do: ", err)
		//return
		os.Exit(1)
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	fmt.Println(timestampMetrics)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return "OK"

}
