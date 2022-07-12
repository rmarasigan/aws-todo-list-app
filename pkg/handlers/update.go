package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/models"
	"github.com/rmarasigan/aws-todo-list-app/pkg/response"
)

// UpdateUser updates and returns the updated account information.
func UpdateUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Get user_id in request header
	user_id := request.Headers["user_id"]

	// Checks if the user_id value is set
	if user_id == "" {
		err := errors.New("user_id is not set")

		logger.Error(err, &logger.Logs{Code: "UserRequestHeaderError", Message: err.Error()})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	user := new(models.User)
	body := []byte(request.Body)

	// Parse the request body of User
	err := response.ParseJSON(body, user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "JSONError", Message: fmt.Sprintf("Failed to parse request body of user %s", user_id)})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}
	user.UserID = user_id

	// Result will return the specific user
	result, err := models.UpdateUserAccount(ctx, user, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: fmt.Sprintf("Failed to update user %s", user_id)})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	// Returns a user response in JSON format
	userResponse, err := models.UserResponse(result)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed umarshal user response"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return api.Response(http.StatusOK, userResponse)
}

// UpdateTask updates and returns the updated task.
func UpdateTask(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request path parameter
	taskID := request.PathParameters["task_id"]

	// Check if taskID is not empty and if not convert taskID string to int
	if taskID == "" {
		return api.UnhandledRequest()
	}

	task := new(models.Task)
	body := []byte(request.Body)

	// Parse the request body of Task
	err := response.ParseJSON(body, task)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "JSONError", Message: fmt.Sprintf("Failed to parse request body of task %s", taskID)})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	// Result will return the specific task
	result, err := models.UpdateTask(ctx, taskID, task, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: fmt.Sprintf("Failed to update task %s", taskID)})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	// Returns a task response in JSON format
	taskResponse, err := models.TaskResponse(result)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed umarshal task response"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return api.Response(http.StatusOK, taskResponse)
}
