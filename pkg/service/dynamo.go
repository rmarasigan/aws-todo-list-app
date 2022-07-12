package service

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	DynamoDBClient dynamodbiface.DynamoDBAPI
)

func DynamoDBInit() dynamodbiface.DynamoDBAPI {
	if DynamoDBClient != nil {
		return DynamoDBClient
	}

	sess := SetSession()
	// Create DynamoDB client
	svc := dynamodb.New(sess)
	// DynamoDBClient = svc
	return svc
}
