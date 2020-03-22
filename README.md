# go-s3_get_files

## Environment variables

To configure and start the program create a shell script (start.sh for example) add the following variables and values:
export TIMEZONE=Europe/Rome
export ESB_AWS_ACCESS_KEY_ID=
export ESB_AWS_SECRET_ACCESS_KEY=
export ESB_AWS_REGION=eu-west-1
export DD_API_KEY=
export DD_APPLICATION_KEY=
export DD_SITE=eu
go run main.go

Spin up the program by running ./start.sh

Inspiration:
https://medium.com/emvi/configuring-golang-applications-using-environment-variables-abf7a76ae506