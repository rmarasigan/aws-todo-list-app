package service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
)

type DynamoBuilder struct {
	TableName     string
	Key           string
	Value         interface{}
	Update        expression.UpdateBuilder
	Filter        expression.ConditionBuilder
	Projection    expression.ProjectionBuilder
	KeyExpression expression.KeyConditionBuilder
	KeyAttribute  map[string]types.AttributeValue
}

func DynamoScan(ctx context.Context, tablename string) (*dynamodb.ScanOutput, error) {
	svc, err := DynamoDBInit(ctx)
	if err != nil {
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String(tablename),
	}

	// Make DynamoDB API Call. Returns one or more items.
	result, err := svc.Scan(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DynamoScanExpression(ctx context.Context, builder DynamoBuilder) (*dynamodb.ScanOutput, error) {
	svc, err := DynamoDBInit(ctx)
	if err != nil {
		return nil, err
	}

	// Building expression with filter and projection to return the specific data and its columns
	expr, err := expression.NewBuilder().WithFilter(builder.Filter).WithProjection(builder.Projection).Build()
	if err != nil {
		return nil, err
	}

	// Build the scan query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(builder.TableName),
	}

	// Make DynamoDB API Call. Returns one or more items.
	result, err := svc.Scan(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DynamoQuery(ctx context.Context, builder DynamoBuilder) (*dynamodb.QueryOutput, error) {
	svc, err := DynamoDBInit(ctx)
	if err != nil {
		return nil, err
	}

	expr, err := expression.NewBuilder().WithKeyCondition(builder.KeyExpression).WithProjection(builder.Projection).Build()
	if err != nil {
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 aws.String(builder.TableName),
	}

	// Make DynamoDB API Call.
	result, err := svc.Query(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DynamoPutItem(ctx context.Context, tablename string, value map[string]types.AttributeValue) (*dynamodb.PutItemOutput, error) {
	svc, err := DynamoDBInit(ctx)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoScan", Message: "Failed to initialize DynamoDB session"}, logger.KVP{Key: "Table", Value: tablename})
		return nil, err
	}

	// Creating the data that you want to send to DynamoDB
	params := &dynamodb.PutItemInput{
		// Map of attribute name-value pairs, one for each attribute
		Item:      value,
		TableName: aws.String(tablename),
	}

	// Creates a new item/object
	result, err := svc.PutItem(ctx, params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoPutItem", Message: "Failed to create item"}, logger.KVP{Key: "Table", Value: tablename})
		return nil, err
	}

	return result, nil
}

func DynamoUpdateItem(ctx context.Context, builder DynamoBuilder) (*dynamodb.UpdateItemOutput, error) {
	svc, err := DynamoDBInit(ctx)
	if err != nil {
		return nil, err
	}

	expr, err := expression.NewBuilder().WithUpdate(builder.Update).Build()
	if err != nil {
		return nil, err
	}

	params := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		Key:                       builder.KeyAttribute,
		ReturnValues:              types.ReturnValueAllNew,
		TableName:                 aws.String(builder.TableName),
	}

	result, err := svc.UpdateItem(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DynamoDeleteItem(ctx context.Context, builder DynamoBuilder) (*dynamodb.DeleteItemOutput, error) {
	svc, err := DynamoDBInit(ctx)
	if err != nil {
		return nil, err
	}

	// Create a delete query based on the parameter
	params := &dynamodb.DeleteItemInput{
		// `Key` is a required field. They key is the primary key that uniquely identifies each item in the table.
		// An AttributeValue represents the data for an attribute.
		Key:       builder.KeyAttribute,
		TableName: aws.String(builder.TableName),
		// Returns all of the attributes of the item (before the DeleteItem operation)
		ReturnValues: types.ReturnValueAllOld,
	}

	result, err := svc.DeleteItem(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}
