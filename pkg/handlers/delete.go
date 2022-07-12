package handlers

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/models"
)

// DeleteUser deletes the specific user and returns a response of the old deleted user details.
func DeleteUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request parameter
	id := request.PathParameters["user_id"]

	// Reulst will return the delete output
	result, err := models.DeleteAccount(ctx, id, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to delete user"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	// Returns a user response in JSON format
	userResponse, err := models.UserResponse(result)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal user response"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return api.Response(http.StatusOK, userResponse)
}

// DeleteTask deletes the specific task and returns a response of the old deleted task details.
func DeleteTask(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request parameter
	id := request.PathParameters["task_id"]

	// Result will return the delete ouput
	result, err := models.DeleteTask(ctx, id, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to delete task"})
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
