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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/service"
)

const (
	BACKLOG     = "Backlog"
	IN_PROGRESS = "In Progress"
	DONE        = "Done"
)

// StatusMap maps all the task status
//
// 1: Backlog || 2: In Progress || 3: Done
var StatusMap = map[int]string{
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
func TaskResponse(result map[string]types.AttributeValue) (*Task, error) {
	task := new(Task)

	// Unmarshal it into actual task which front-end can understand as a JSON
	err := attributevalue.UnmarshalMap(result, task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "TaskResponse", Message: "Failed to unmarshal result map to task"})
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
func GetTask(ctx context.Context, task_id string) (*Task, error) {
	// Initialize a struct and returns a pointer to an instance of Task struct
	task := new(Task)
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("TASKS_TABLE")

	// Create the expression to fill the input struct
	// KeyConditionExpression: "id = task_id_value"
	builder.KeyExpression = expression.Key("id").Equal(expression.Value(task_id))

	// ProjectionExpression: id, title, description, status, date_created
	builder.Projection = expression.NamesList(expression.Name("id"), expression.Name("title"), expression.Name("description"), expression.Name("status"), expression.Name("date_created"))

	// Make DynamoDB query API Call
	result, err := service.DynamoQuery(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetTask", Message: "Failed to execute DynamoQuery"}, logger.KVP{Key: "task_id", Value: task_id})
		return nil, err
	}

	// Checks if the result length is 0 to return a message of "No data found"
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "GetTask", Message: "No data found"}, logger.KVP{Key: "task_id", Value: task_id})
		return nil, nil
	}

	// Unmarshal it into actual task which front-end can understand as a JSON
	err = attributevalue.UnmarshalMap(result.Items[0], task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetTask", Message: "Failed to unmarshal task record"})
		return nil, err
	}

	return task, nil
}

// GenerateTaskID generates a new ID if the current ID argument passed already exist.
func GenerateTaskID(ctx context.Context, id int) (string, error) {
	var tasks []Task
	tablename := os.Getenv("TASKS_TABLE")

	result, err := service.DynamoScan(ctx, tablename)
	if err != nil {
		return "", err
	}

	// Unmarshal it into actual task for us to iterate
	err = attributevalue.UnmarshalListOfMaps(result.Items, &tasks)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(tasks); i++ {
		taskID, err := strconv.Atoi(tasks[i].TaskID)
		if err != nil {
			return "", err
		}

		// Check if the current ID already exist
		if id == taskID {
			// Loop backwards to check if the ID exist
			for j := (len(tasks) - 1); j >= 0; j-- {
				_taskID, err := strconv.Atoi(tasks[j].TaskID)
				if err != nil {
					return "", err
				}

				// Check backward if the current ID already exist
				if id == _taskID {
					i = 0
					id -= 1
					continue
				}
			}
		}
	}

	return fmt.Sprint(id), nil
}

// CreateTask creates a new object of task that will be saved on dynamoDB table.
func CreateTask(ctx context.Context, task *Task) (*events.APIGatewayProxyResponse, error) {
	tablename := os.Getenv("TASKS_TABLE")

	// Get the total count of tasks to set TaskID
	count := ItemCount(ctx, tablename) + 1

	// Checks if ID exist and generate a new one
	id, err := GenerateTaskID(ctx, count)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "CreateTask", Message: "Failed to GenerateTaskID"})
		return api.StatusBadRequest(err)
	}
	task.TaskID = id

	// Convert status string to int
	taskStatus, err := strconv.Atoi(task.Status)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "CreateTask", Message: "Failed to convert task status string to integer"})
		return api.StatusBadRequest(err)
	}

	// Set task status value
	task.Status = StatusMap[taskStatus]

	// Converting the record to dynamodb.AttributeValue type
	value, err := attributevalue.MarshalMap(task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "CreateTask", Message: "Failed to marshal task"})
		return api.StatusBadRequest(err)
	}

	// Validate required fields
	validate := task.Validate()
	if validate != "" {
		err := errors.New(validate)
		logger.Error(err, &logger.Logs{Code: "CreateTask", Message: "Required fields are not entered"})

		return api.StatusBadRequest(err)
	}

	// Creates a new item
	_, err = service.DynamoPutItem(ctx, tablename, value)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "CreateTask", Message: "Failed to execute DynamoPutItem"})
		return api.StatusBadRequest(err)
	}

	return api.Response(http.StatusOK, task)
}

// FetchTasks returns list of tasks of the said user.
func FetchTasks(ctx context.Context, user_id string) (*[]Task, error) {
	// Initialize a struct and returns a pointer to an instance of Task struct
	tasks := new([]Task)
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("TASKS_TABLE")

	// Create the expression to fill the input struct
	// FilterExpression: "user_id = user_id_value"
	builder.Filter = expression.Name("user_id").Equal(expression.Value(user_id))

	// ProjectionExpression: id, title, description, status, date_created
	builder.Projection = expression.NamesList(expression.Name("id"), expression.Name("title"), expression.Name("description"), expression.Name("status"), expression.Name("date_created"))

	// Make DynamoDB API Call. Returns one or more items.
	result, err := service.DynamoScanExpression(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "FetchTasks", Message: "Failed to execute DynamoScanExpression"})
		return nil, err
	}

	// Checks if there are items returned
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "FetchTasks", Message: "No data found"})
		return nil, nil
	}

	// Unmarshal a list of maps into actual task which front-end can understand as a JSON
	err = attributevalue.UnmarshalListOfMaps(result.Items, tasks)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "FetchTasks", Message: "Failed to unmarshal tasks record"})
		return nil, err
	}

	return tasks, nil
}

// FilterTasks returns a list of tasks of the said user depending on the status or progress of the Task.
func FilterTasks(ctx context.Context, user_id string, status int) (*[]Task, error) {
	// Initialize a struct and returns a pointer to an instance of Task struct
	tasks := new([]Task)
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("TASKS_TABLE")

	// Create the expression to fill the input struct
	// FilterExpression: "user_id = user_id_value AND status = status_value"
	builder.Filter = expression.Name("user_id").Equal(expression.Value(user_id)).And(expression.Name("status").Equal(expression.Value(StatusMap[status])))

	// ProjectionExpression: id, title, description, status, date_created
	builder.Projection = expression.NamesList(expression.Name("id"), expression.Name("title"), expression.Name("description"), expression.Name("status"), expression.Name("date_created"))

	// Make DynamoDB query API Call. Returns one or more.
	result, err := service.DynamoScanExpression(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "FilterTasks", Message: "Failed to execute DynamoScanExpression"}, logger.KVP{Key: "user_id", Value: user_id}, logger.KVP{Key: "status", Value: status})
		return nil, err
	}

	// Checks if there are items returned
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "FilterTasks", Message: "No data found"})
		return nil, nil
	}

	// Unmarshal a list of maps into actual task which front-end can understand as a JSON
	err = attributevalue.UnmarshalListOfMaps(result.Items, tasks)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "FilterTasks", Message: "Failed to unmashal tasks record"}, logger.KVP{Key: "user_id", Value: user_id}, logger.KVP{Key: "status", Value: status})
		return nil, err
	}

	return tasks, nil
}

// UpdateTask will update the specific Task information.
func UpdateTask(ctx context.Context, id string, task *Task) (map[string]types.AttributeValue, error) {
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("TASKS_TABLE")

	// Get the specific task
	taskExist, err := GetTask(ctx, id)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "UpdateTask", Message: "Failed to execute GetTask"}, logger.KVP{Key: "task_id", Value: id})
		return nil, err
	}

	// Check if the task exist before updating
	if taskExist == nil {
		logger.Info(&logger.Logs{Code: "UpdateTask", Message: "The task does not exist"}, logger.KVP{Key: "task_id", Value: id})
		return nil, nil
	}

	// Validate if fields are empty
	task = task.ValidateEmpty(taskExist)

	// Convert status string to int
	taskStatus, err := strconv.Atoi(task.Status)
	if err != nil {
		err := errors.New("failed to convert status to int")
		logger.Error(err, &logger.Logs{Code: "UpdateTask", Message: "Cannot convert task status string to int"}, logger.KVP{Key: "task_id", Value: id})

		return nil, err
	}

	// Set task status value
	task.Status = StatusMap[taskStatus]

	builder.KeyAttribute = map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{
			Value: id,
		},
	}

	// Set a name-value pair for update expression
	builder.Update = expression.Set(expression.Name("title"), expression.Value(task.Title)).
		Set(expression.Name("description"), expression.Value(task.Description)).
		Set(expression.Name("status"), expression.Value(task.Status)).
		Set(expression.Name("date_updated"), expression.Value(task.SetCurrentDateTime()))

	// Updates the specific field
	result, err := service.DynamoUpdateItem(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "UpdateTask", Message: "Failed to execute DynamoUpdateItem"})
		return nil, err
	}

	return result.Attributes, nil
}

// DeleteTask deletes a specific task based on the id given.
func DeleteTask(ctx context.Context, id string) (map[string]types.AttributeValue, error) {
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("TASKS_TABLE")

	// Get the specific task
	taskExist, err := GetTask(ctx, id)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteTask", Message: "Failed to get task"})
		return nil, err
	}

	// Check if the task exist before deleting
	if taskExist == nil {
		err := errors.New("task does not exist")
		logger.Error(err, &logger.Logs{Code: "DeleteTask", Message: "The task does not exist"})

		return nil, err
	}

	builder.KeyAttribute = map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{
			Value: id,
		},
	}

	// Deletes a single item in a table by primary key
	result, err := service.DynamoDeleteItem(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteTask", Message: "Failed to delete task"}, logger.KVP{Key: "task_id", Value: id})
		return nil, err
	}

	return result.Attributes, nil
}

// DeleteUserTasks deletes all tasks related to the user account that is being deleted.
func DeleteUserTasks(ctx context.Context, user_id string) (*events.APIGatewayProxyResponse, error) {
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("TASKS_TABLE")

	tasks, err := FetchTasks(ctx, user_id)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteUserTasks", Message: "Failed to fetch all tasks"})
		return api.StatusBadRequest(err)
	}

	// Loop over the tasks the user has and delete one by one
	for _, task := range *tasks {
		builder.KeyAttribute = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: task.TaskID,
			},
		}

		// Deletes all tasks related to the user account
		_, err = service.DynamoDeleteItem(ctx, builder)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "DeleteUserTasks", Message: "Failed to delete all tasks"})
			return api.StatusBadRequest(err)
		}
	}

	return api.Response(http.StatusOK, "all tasks related to the account are deleted")
}
