package api

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rmarasigan/aws-todo-list-app/pkg/response"
)

type Error struct {
	Message *string `json:"err_msg,omitempty"`
}

// Response returns a response to be returned by the API Gateway Request.
func Response(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                  "application/json",
			"Access-Control-Allow-Origin":   "*",
			"Access-Control-Request-Method": "*",
			"Access-Control-Allow-Headers":  "Content-Type,X-Api-Key",
		},
		StatusCode: status,
		Body:       response.EncodeResponseJSON(status, body),
	}, nil
}

// StatusBadRequest returns a response of an HTTP StatusBadRequest and an error message
func StatusBadRequest(err error) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                  "application/json",
			"Access-Control-Allow-Origin":   "*",
			"Access-Control-Request-Method": "*",
			"Access-Control-Allow-Headers":  "Content-Type,X-Api-Key",
		},
		StatusCode: http.StatusBadRequest,
		Body:       response.EncodeJSON(Error{Message: aws.String(err.Error())}),
	}, nil
}

// UnhandledMethod returns an http status code "StatusMethodNotAllowed" and an error of "method not allowed"
func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return Response(http.StatusMethodNotAllowed, "method not allowed")
}

// UnhandledRequest returns an http status code "StatusBadRequest" and an error of "request not allowed"
func UnhandledRequest() (*events.APIGatewayProxyResponse, error) {
	return Response(http.StatusBadRequest, "request not allowed")
}
