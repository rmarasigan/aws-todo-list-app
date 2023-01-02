package handlers

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/models"
)

// DeleteUser deletes the specific user and returns a response of the old deleted user details.
func DeleteUser(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request parameter
	id := request.PathParameters["user_id"]

	// Reulst will return the delete output
	result, err := models.DeleteAccount(ctx, id)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteUser", Message: "Failed to delete user"}, logger.KVP{Key: "user_id", Value: id})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	// Returns a user response in JSON format
	userResponse, err := models.UserResponse(result)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteUser", Message: "Failed to unmarshal user response"})
		return api.Response(http.StatusBadRequest, api.Error{Message: aws.String(err.Error())})
	}

	return api.Response(http.StatusOK, userResponse)
}

// DeleteTask deletes the specific task and returns a response of the old deleted task details or deletes all tasks
// of the user if the user deleted his/her account.
func DeleteTask(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Gets the request parameter
	task_id := request.PathParameters["task_id"]

	// Get the request header
	user_id := request.Headers["user_id"]

	// Checks if there is a selected task
	if len(task_id) > 0 {
		// Result will return the delete ouput
		result, err := models.DeleteTask(ctx, task_id)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "DeleteTask", Message: "Failed to delete task"}, logger.KVP{Key: "task_id", Value: task_id})
			return api.StatusBadRequest(err)
		}

		// Returns a task response in JSON format
		taskResponse, err := models.TaskResponse(result)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "DeleteTask", Message: "Failed umarshal task response"})
			return api.StatusBadRequest(err)
		}

		return api.Response(http.StatusOK, taskResponse)
	} else {

		// Deletes the user tasks if the user deleted the account
		result, err := models.DeleteUserTasks(ctx, user_id)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "DeleteTask", Message: "Failed to delete user tasks"}, logger.KVP{Key: "user_id", Value: user_id})
			return api.StatusBadRequest(err)
		}

		return result, nil
	}
}
