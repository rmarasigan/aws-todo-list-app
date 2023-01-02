package service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
)

var (
	AWS_REGION = "us-east-1"
)

// InitSession initialize the SDK configuration
func InitSession(ctx context.Context) (cfg aws.Config, err error) {
	// Creates a new session
	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(AWS_REGION))
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "InitSession", Message: "Failed to create a new AWS session"})
		return
	}

	return
}

// DynamoDBInit creates a new instance of the DynamoDB client with a session
func DynamoDBInit(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := InitSession(ctx)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBInit", Message: "Failed to laod SDK default configuration"})
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}
