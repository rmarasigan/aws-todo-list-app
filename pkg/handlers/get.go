package handlers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/models"
	"github.com/rmarasigan/aws-todo-list-app/pkg/response"
)

// GetUser gets the specifc user information
func GetUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tablename := os.Getenv("USERS_TABLE")

	// Gets the request parameter
	username := request.QueryStringParameters["username"]

	if username == "" {
		return api.UnhandledRequest()
	}

	// Result will return the specific user
	result, err := models.FetchUser(ctx, username, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to fetch user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
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
		logger.Error(err, &logger.Logs{Code: "JSONError", Message: "Failed to parse request body of user"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	result, err := models.UserAuthentication(ctx, user, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBEror", Message: "Failed to authenticate user"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return api.Response(http.StatusOK, result)
}

// GetTask gets a Task(s) based on the status parameter if set. If not, it will return all list of Tasks for the said user.
func GetTask(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request paramater
	status := request.QueryStringParameters["status"]

	// Checks if the user_id value is set
	user_id := request.Headers["user_id"]
	if user_id == "" {
		err := errors.New("user_id is not set")

		logger.Error(err, &logger.Logs{Code: "TaskRequestHeaderError", Message: err.Error()})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	if len(status) > 0 {
		// Convert status string to int
		taskStatus, err := strconv.Atoi(status)
		if err != nil {
			return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
		}

		// Result will return a list of tasks base on status
		result, err := models.FilterTasks(ctx, user_id, taskStatus, svc)
		if err != nil {
			return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
		}

		return api.Response(http.StatusOK, result)
	}

	// Result will return the whole list of tasks
	result, err := models.FetchTasks(ctx, user_id, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to fetch tasks"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return api.Response(http.StatusOK, result)
}
