package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
)

const (
	BACKLOG     = "Backlog"
	IN_PROGRESS = "In Progress"
	DONE        = "Done"
)

// statusMap maps all the task status
//
// 1: Backlog || 2: In Progress || 3: Done
var statusMap = map[int]string{
	1: BACKLOG,
	2: IN_PROGRESS,
	3: DONE,
}

// Task contains the task details of the client
type Task struct {
	TaskID      string `json:"id" dynamodbav:"id"`
	UserID      string `json:"user_id" dynamodbav:"user_id"`
	Title       string `json:"title" dynamodbav:"title"`
	Description string `json:"description" dynamodbav:"description"`
	Status      string `json:"status" dynamodbav:"status"`
	DateCreated string `json:"date_created" dynamodbav:"date_created"`
	DateUpdated string `json:"date_updated" dynamodbav:"date_updated"`
}

// SetCurrentDateTime formats the current date and time
func (t *Task) SetCurrentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// TaskResponse returns a Task struct with the values of result object.
func TaskResponse(result map[string]*dynamodb.AttributeValue) (*Task, error) {
	task := new(Task)

	// Unmarshal it into actual task which front-end can understand as a JSON
	err := dynamodbattribute.UnmarshalMap(result, task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal result map to task"})
		return nil, err
	}

	return task, nil
}

// ValidateEmpty validates if the field is empty or not to set its previous value.
func (task *Task) ValidateEmpty(old *Task) *Task {
	if task.Title == "" {
		task.Title = old.Title
	}

	if task.Description == "" {
		task.Description = old.Description
	}

	if task.Status == "" {
		task.Status = old.Status
	}

	return task
}

// Validate validates the required field.
func (task *Task) Validate() string {
	var msg []string
	var err_msg string

	if task.Title == "" {
		msg = append(msg, "Title")
	}

	if task.Status == "" {
		msg = append(msg, "Status")
	}

	if len(msg) != 0 {
		err_msg = fmt.Sprintf("Missing %s field(s)", strings.Join(msg, ", "))
	}

	return err_msg
}

// GetTask gets a specifc task using the id parameter.
func GetTask(ctx context.Context, task_id string, svc dynamodbiface.DynamoDBAPI) (*Task, error) {
	// Initialize a struct and returns a pointer to an instance of Task struct
	task := new(Task)
	tablename := os.Getenv("TASKS_TABLE")

	// Create the expression to fill the input struct
	// KeyConditionExpression: "id = task_id_value"
	key := expression.Key("id").Equal(expression.Value(task_id))

	// ProjectionExpression: id, title, description, status, date_created
	projection := expression.NamesList(expression.Name("id"), expression.Name("title"), expression.Name("description"), expression.Name("status"), expression.Name("date_created"))

	// SELECT id, title, description, status, date_created WHERE id = task_id_value
	expr, err := expression.NewBuilder().WithKeyCondition(key).WithProjection(projection).Build()
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "ExpressionError", Message: "Failed to create expression"},
			logger.KVP{Key: "tablename", Value: aws.String(tablename)})
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.QueryInput{
		TableName:                 aws.String(tablename),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	// Make DynamoDB query API Call
	result, err := svc.QueryWithContext(ctx, params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBAPIError", Message: "Failed to call dynamoDB query API"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Checks if the result length is 0 to return a message of "No data found"
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "TaskData", Message: "No data found"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, nil
	}

	// Unmarshal it into actual task which front-end can understand as a JSON
	err = dynamodbattribute.UnmarshalMap(result.Items[0], task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal task record"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return task, nil
}

// CreateTask creates a new object of task that will be saved on dynamoDB table.
func CreateTask(ctx context.Context, task *Task, svc dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
	tablename := os.Getenv("TASKS_TABLE")

	// Get the total count of tasks to set TaskID
	items := ItemCount(tablename, svc)
	task.TaskID = fmt.Sprint(1 + items)

	// Convert status string to int
	taskStatus, err := strconv.Atoi(task.Status)
	if err != nil {
		return api.StatusBadRequest(err)
	}

	// Set task status value
	task.Status = statusMap[taskStatus]

	// Converting the record to dynamodb.AttributeValue type
	value, err := dynamodbattribute.MarshalMap(task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to marshal task"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Validate required fields
	validate := task.Validate()
	if validate != "" {
		err := errors.New(validate)

		logger.Error(err, &logger.Logs{Code: "NewTaskValidation", Message: "Required fields are not entered"},
			logger.KVP{Key: "validation", Value: validate})
		return nil, err
	}

	// Creating the data that you want to send to dynamoDB
	params := &dynamodb.PutItemInput{
		// Map of attribute name-value pairs, one for each attribute
		Item:      value,
		TableName: aws.String(tablename),
	}

	// Creates a new item
	_, err = svc.PutItem(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to create a new task"})
		return nil, err
	}

	return api.Response(http.StatusOK, task)
}

// FetchTasks returns list of tasks of the said user.
func FetchTasks(ctx context.Context, user_id string, svc dynamodbiface.DynamoDBAPI) (*[]Task, error) {
	// Initialize a struct and returns a pointer to an instance of Task struct
	tasks := new([]Task)
	tablename := os.Getenv("TASKS_TABLE")

	// Create the expression to fill the input struct
	// FilterExpression: "user_id = user_id_value"
	filter := expression.Name("user_id").Equal(expression.Value(user_id))

	// ProjectionExpression: id, title, description, status, date_created
	projection := expression.NamesList(expression.Name("id"), expression.Name("title"), expression.Name("description"), expression.Name("status"), expression.Name("date_created"))

	// SELECT id, title, description, status, date_created FROM tablename WHERE user_id = 'user_id_value'
	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(projection).Build()
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "ExpressionError", Message: "Failed to create expression"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName:                 aws.String(tablename),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	}

	// Make DynamoDB query API Call. Returns one or more items and item attributes by accessing item in a table or a secondary index
	result, err := svc.Scan(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBAPIError", Message: "Failed to call dynamoDB Query API"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Checks if there are items returned
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "DynamoDBAPI", Message: "No data found"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, nil
	}

	// Unmarshal a list of maps into actual task which front-end can understand as a JSON
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, tasks)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal tasks record"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return tasks, nil
}

// FilterTasks returns a list of tasks of the said user depending on the status or progress of the Task
func FilterTasks(ctx context.Context, user_id string, status int, svc dynamodbiface.DynamoDBAPI) (*[]Task, error) {
	// Initialize a struct and returns a pointer to an instance of Task struct
	tasks := new([]Task)
	tablename := os.Getenv("TASKS_TABLE")

	// Create the expression to fill the input struct
	// FilterExpression: "user_id = user_id_value AND status = status_value"
	filter := expression.Name("user_id").Equal(expression.Value(user_id)).And(expression.Name("status").Equal(expression.Value(statusMap[status])))

	// ProjectionExpression: id, title, description, status, date_created
	projection := expression.NamesList(expression.Name("id"), expression.Name("title"), expression.Name("description"), expression.Name("status"), expression.Name("date_created"))

	// SELECT user_id, title, description, status, date_created FROM tablename WHERE status = 'status_value'
	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(projection).Build()
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "ExpressionError", Message: "Failed to create expression"},
			logger.KVP{Key: "tablename", Value: aws.String(tablename)})
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName:                 aws.String(tablename),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	}

	// Make DynamoDB query API Call. Returns one or more items and item attributes by accessing item in a table or a secondary index
	result, err := svc.Scan(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBAPIError", Message: "Failed to call dynamoDB Query API"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Checks if there are items returned
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "DynamoDBAPI", Message: "No data found"}, logger.KVP{Key: "tablename", Value: tablename})
		return nil, nil
	}

	// Unmarshal a list of maps into actual task which front-end can understand as a JSON
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, tasks)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmashal tasks record"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return tasks, nil
}

// UpdateTask will update the specific Task information.
func UpdateTask(ctx context.Context, id string, task *Task, svc dynamodbiface.DynamoDBAPI) (map[string]*dynamodb.AttributeValue, error) {
	tablename := os.Getenv("TASKS_TABLE")

	// Get the specific task
	taskExist, err := GetTask(ctx, id, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to get task"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Check if the task exist before updating
	if taskExist == nil {
		logger.Info(&logger.Logs{Code: "UpdateTaskData", Message: "The task does not exist"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, errors.New("task does not exist")
	}

	// Validate if fields are empty
	task = task.ValidateEmpty(taskExist)

	// Set a name-value pair for update expression
	update := expression.Set(expression.Name("title"), expression.Value(task.Title)).
		Set(expression.Name("description"), expression.Value(task.Description)).
		Set(expression.Name("status"), expression.Value(task.Status)).
		Set(expression.Name("date_updated"), expression.Value(task.SetCurrentDateTime()))

	// Build the expression.
	// UPDATE tablename SET title = title_value, description = description_value, ... WHERE id = id_value
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "ExpressionError", Message: "Failed to create expression"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Build the update item input parameters
	params := &dynamodb.UpdateItemInput{
		TableName:                 &tablename,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		// Returns all of the attributes of the item (after the UpdateItem operation)
		ReturnValues: aws.String("ALL_NEW"),
		// `Key` is a required field. They key is the primary key that uniquely identifies each item in the table.
		// An AttributeValue represents the data for an attribute.
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: &id},
		},
	}

	// Updates the specific field
	output, err := svc.UpdateItemWithContext(ctx, params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to update task"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return output.Attributes, nil
}

// DeleteTask deletes a specific task based on the id given.
func DeleteTask(ctx context.Context, id string, svc dynamodbiface.DynamoDBAPI) (map[string]*dynamodb.AttributeValue, error) {
	tablename := os.Getenv("TASKS_TABLE")

	// Get the specific task
	taskExist, err := GetTask(ctx, id, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to get task"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Check if the task exist before deleting
	if taskExist == nil {
		err := errors.New("task does not exist")

		logger.Error(err, &logger.Logs{Code: "DeleteTaskData", Message: "The task does not exist"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Create a delete query based on the parameter
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(tablename),
		// Returns all of the attributes of the item (before the DeleteItem operation)
		ReturnValues: aws.String("ALL_OLD"),
		// `Key` is a required field. They key is the primary key that uniquely identifies each item in the table.
		// An AttributeValue represents the data for an attribute.
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	}

	// Deletes a single item in a table by primary key
	output, err := svc.DeleteItem(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to delete task"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return output.Attributes, nil
}
