package main

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rmarasigan/aws-todo-list-app/pkg/api"
	"github.com/rmarasigan/aws-todo-list-app/pkg/handlers"
)

// Entry point that runs your Lambda function code
func main() {
	// It will execute our Lambda function
	lambda.Start(handler)
}

// Main entry function
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var path = request.Path

	// Removes leading slash
	path = strings.TrimPrefix(path, "/")

	// Specifies the integration's HTTP method type
	var http_method = request.HTTPMethod

	switch {
	case strings.HasPrefix(path, "tasks"):
		return TaskHandleRequest(ctx, http_method, request)

	case strings.HasPrefix(path, "users"):
		return UserHandleRequest(ctx, http_method, request)

	default:
		return api.UnhandledMethod()
	}
}

// UserHandleRequest handles all request coming from users
func UserHandleRequest(ctx context.Context, method string, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch method {
	case "GET":
		return handlers.GetUser(ctx, request)

	case "POST":
		// Gets the request parameter
		requestType := request.QueryStringParameters["type"]

		// Check request type if login, create and update
		switch requestType {
		case "login":
			return handlers.LoginUser(ctx, request)

		case "new":
			return handlers.CreateUser(ctx, request)

		case "update":
			return handlers.UpdateUser(ctx, request)

		default:
			return api.UnhandledMethod()
		}

	case "DELETE":
		return handlers.DeleteUser(ctx, request)

	default:
		return api.UnhandledMethod()
	}
}

// TaskHandleRequest handles all request coming from tasks.
func TaskHandleRequest(ctx context.Context, method string, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch method {
	case "GET":
		return handlers.GetTask(ctx, request)

	case "POST":
		// Gets the request parameter
		requestType := request.QueryStringParameters["type"]

		// Check request type to handle create and update
		switch requestType {
		case "new":
			return handlers.CreateTask(ctx, request)

		case "update":
			return handlers.UpdateTask(ctx, request)

		default:
			return api.UnhandledRequest()
		}

	case "DELETE":
		return handlers.DeleteTask(ctx, request)

	default:
		return api.UnhandledMethod()
	}
}
