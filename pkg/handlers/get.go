package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/models"
	"github.com/rmarasigan/aws-todo-list-app/pkg/response"
)

// GetUser gets the specifc user information
func GetUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request parameter
	username := request.QueryStringParameters["username"]

	if username == "" {
		return api.UnhandledRequest()
	}

	// Result will return the specific user
	result, err := models.FetchUser(ctx, username)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetUser", Message: "Failed to fetch user"}, logger.KVP{Key: "username", Value: username})
		return api.StatusBadRequest(err)
	}

	return api.Response(http.StatusOK, result)
}

// LoginUser authenticates the user if the account exist and confirming credentials.
func LoginUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	user := new(models.User)
	body := []byte(request.Body)

	// Parse the request body of User
	err := response.ParseJSON(body, user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "LoginUser", Message: "Failed to parse request body of user"})
		return api.StatusBadRequest(err)
	}

	result, err := models.UserAuthentication(ctx, user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "LoginUser", Message: "Failed to authenticate user"})
		return api.StatusBadRequest(err)
	}

	return api.Response(http.StatusOK, result)
}

// GetTask gets a Task(s) based on the status parameter if set. If not, it will return all list of Tasks for the said user.
func GetTask(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request paramater
	status := request.QueryStringParameters["status"]

	// Gets the request path parameter
	taskID := request.PathParameters["task_id"]

	// Checks if the user_id value is set
	user_id := request.Headers["user_id"]
	if user_id == "" {
		err := errors.New("user_id is not set")
		logger.Error(err, &logger.Logs{Code: "GetTask", Message: err.Error()})

		return api.StatusBadRequest(err)
	}

	// Get tasks with specific status
	if len(status) > 0 {
		// Convert status string to int
		taskStatus, err := strconv.Atoi(status)
		if err != nil {
			return api.StatusBadRequest(err)
		}

		// Result will return a list of tasks base on status
		result, err := models.FilterTasks(ctx, user_id, taskStatus)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "GetTask", Message: "Failed to FilterTasks"},
				logger.KVP{Key: "user_id", Value: user_id}, logger.KVP{Key: "status", Value: models.StatusMap[taskStatus]})
			return api.StatusBadRequest(err)
		}

		return api.Response(http.StatusOK, result)
	}

	// Get specific task information
	if len(taskID) > 0 {
		// Result will return the specific task details
		result, err := models.GetTask(ctx, taskID)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "GetTask", Message: "Failed to get task"}, logger.KVP{Key: "task_id", Value: taskID})
			return api.StatusBadRequest(err)
		}

		return api.Response(http.StatusOK, result)
	}

	// Result will return the whole list of tasks
	result, err := models.FetchTasks(ctx, user_id)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetTask", Message: "Failed to fetch tasks"}, logger.KVP{Key: "user_id", Value: user_id})
		return api.StatusBadRequest(err)
	}

	return api.Response(http.StatusOK, result)
}
