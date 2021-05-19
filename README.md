# go-s3_get_files

## Introduction

List content of a AWS S3 bucket or prefix, filter out the oldest S3 object based on the LastModified attribute and send the age in seconds of the object as a metric to Datadog. Monitoring is established to be notified when objects are left behind and aren’t picked up within a certain time frame.

## Implementation highlights

Pseudo Golang code:
- list S3 bucket content;
- for the first item extract the LastModified attribute value and store it together with the object name in a temporary variable;
- for each consecutive item extract the LastModified attribute value and compare it with the already value stored in the temporary variable. If the object is older then overwrite it with the current item meta data. The temporary variable will contain the oldest item;
- The object age (in seconds) and meta data of the stored item will then be communicated to DataDog through the DataDog API.

A DataDog Monitor per S3 bucket sends out alerts when configured thresholds are exceeded. 

Per S3 bucket (or prefix), deploy an AWS Lambda function with its own scheduling configuration using the CloudWatch Event service from AWS.

## Environment variables

To configure and start the program create a shell script (start.sh for example) add the following variables and values:
```console
export TIMEZONE=Europe/Rome\
export ESB_AWS_ACCESS_KEY_ID=\
export ESB_AWS_SECRET_ACCESS_KEY=\
export ESB_AWS_REGION=eu-west-1\
export DD_API_KEY=\
export DD_APPLICATION_KEY=\
export DD_SITE=eu\
go run main.go\
```
\
Spin up the program by running ./start.sh\

## Build and package

Build with
```console
GOOS=linux GOARCH=amd64 go build -o main main.go
```

Then package with
```console
zip main.zip main
```

Inspiration:\
https://medium.com/emvi/configuring-golang-applications-using-environment-variables-abf7a76ae506
