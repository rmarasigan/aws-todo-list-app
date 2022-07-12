package handlers

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/rmarasigan/aws-todo-list-app/pkg/service"
)

var (
	svc dynamodbiface.DynamoDBAPI
)

func init() {
	svc = service.DynamoDBInit()
}
