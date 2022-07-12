package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
)

var sess *session.Session
var region string

func SetSession() *session.Session {
	if sess != nil {
		return sess
	}

	// Creates a new session
	newSession, err := session.NewSession(&aws.Config{})
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "SessionError", Message: "Failed to create a new AWS Session"})
		panic(err)
	}

	sess = newSession

	// Set Region
	region = *sess.Config.Region
	if region == "" {
		sess.Config.Region = setRegion()
	}

	return sess
}

func setRegion() *string {
	return aws.String("us-east-1")
}
