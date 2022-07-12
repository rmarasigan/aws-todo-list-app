package models

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
)

// ItemCount returns the number of items in your table.
func ItemCount(tablename string, svc dynamodbiface.DynamoDBAPI) int {
	// Build the query input parameters
	params := &dynamodb.ScanInput{TableName: aws.String(tablename)}

	// Make DynamoDB query API Call. Returns one or more items and item attributes by accessing item in a table or a secondary index
	result, err := svc.Scan(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBAPIError", Message: "Failed to call dynamoDB Query API"})
		return 0
	}

	return int(*result.Count)
}
