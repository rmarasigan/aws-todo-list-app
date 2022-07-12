package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/models"

	"github.com/rmarasigan/aws-todo-list-app/pkg/response"
)

// CreateUser creates a new user account and save to dynamodb table.
func CreateUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	user := new(models.User)
	body := []byte(request.Body)

	// Parse the request body of User
	err := response.ParseJSON(body, user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "JSONError", Message: "Failed to parse request body of user"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}
	user.DateCreated = user.SetCurrentDateTime()

	result, err := models.NewUserAccount(ctx, user, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to create user"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return result, nil
}

// CreateTask creates a new task and save to dynamodb table.
func CreateTask(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	task := new(models.Task)
	body := []byte(request.Body)

	// Parse the request body of Task
	err := response.ParseJSON(body, task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "JSONError", Message: "Failed to parse request body of task"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	// Checks if the user_id value is set
	user_id := request.Headers["user_id"]
	if user_id == "" {
		err := errors.New("user_id is not set")

		logger.Error(err, &logger.Logs{Code: "TaskRequestHeaderError", Message: err.Error()})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}
	task.UserID = user_id
	task.DateCreated = task.SetCurrentDateTime()

	result, err := models.CreateTask(ctx, task, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to create task"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return result, nil
}
