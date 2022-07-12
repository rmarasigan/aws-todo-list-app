package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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

// User contains the client account information
type User struct {
	UserID      string `json:"id" dynamodbav:"id"`
	FirstName   string `json:"first_name" dynamodbav:"first_name"`
	LastName    string `json:"last_name" dynamodbav:"last_name"`
	Username    string `json:"username" dynamodbav:"username"`
	Password    string `json:"password" dynamodbav:"password"`
	Email       string `json:"email,omitempty" dynamodbav:"email,omitempty"`
	DateCreated string `json:"date_created" dynamodbav:"date_created"`
}

type Response struct {
	UserID    string `json:"id" dynamodbav:"id"`
	FirstName string `json:"first_name" dynamodbav:"first_name"`
	LastName  string `json:"last_name" dynamodbav:"last_name"`
	Username  string `json:"username" dynamodbav:"username"`
	Email     string `json:"email,omitempty" dynamodbav:"email,omitempty"`
}

// SetCurrentDateTime formats the current date and time.
func (u *User) SetCurrentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// ValidateEmpty validates if the field is empty or not to set its previous value.
func (user *User) ValidateEmpty(old *User) *User {
	if user.FirstName == "" {
		user.FirstName = old.FirstName
	}

	if user.LastName == "" {
		user.LastName = old.LastName
	}

	if user.Username == "" {
		user.Username = old.Username
	}

	if user.Password == "" {
		user.Password = old.Password
	}

	return user
}

// Validate validates the required field.
func (user *User) Validate() string {
	var msg []string
	var err_msg string

	if user.FirstName == "" {
		msg = append(msg, "First Name")
	}

	if user.LastName == "" {
		msg = append(msg, "Last Name")
	}

	if user.Username == "" {
		msg = append(msg, "Username")
	}

	if user.Password == "" {
		msg = append(msg, "Password")
	}

	if user.Email == "" {
		msg = append(msg, "Email")
	}

	if len(msg) != 0 {
		err_msg = fmt.Sprintf("Missing %s field(s)", strings.Join(msg, ", "))
	}

	return err_msg
}

// ValidateAuhtentication validate user authentication request fields.
func (user *User) ValidateAuhtentication() string {
	var msg []string
	var err_msg string

	if user.Username == "" {
		msg = append(msg, "Username")
	}

	if user.Password == "" {
		msg = append(msg, "Password")
	}

	if len(msg) != 0 {
		err_msg = fmt.Sprintf("Missing %s field(s).", strings.Join(msg, ", "))
	}

	return err_msg
}

// UserResponse returns a User struct with the values of result object.
func UserResponse(result map[string]*dynamodb.AttributeValue) (*Response, error) {
	user := new(Response)

	// Unmarshal it into actual task which front-end can understand as a JSON
	err := dynamodbattribute.UnmarshalMap(result, user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal result map to user"})
		return nil, err
	}

	return user, err
}

// GetUser gets a specific user using the id parameter
func GetUser(ctx context.Context, id string, svc dynamodbiface.DynamoDBAPI) (*User, error) {
	// Initialze a struct and returns a pointer to an instance of User struct
	user := new(User)
	tablename := os.Getenv("USERS_TABLE")

	// Create the expression to fill the input struct
	// KeyConditionExpression: "id = user_id_value"
	key := expression.Key("id").Equal(expression.Value(id))

	// ProjectionExpression: id, first_name, last_name, username, email
	projection := expression.NamesList(expression.Name("id"), expression.Name("first_name"), expression.Name("last_name"), expression.Name("username"), expression.Name("email"), expression.Name("password"))

	// SELECT id, first_name, last_name, username, email
	expr, err := expression.NewBuilder().WithKeyCondition(key).WithProjection(projection).Build()
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "ExpressionError", Message: "Faield to create expression"},
			logger.KVP{Key: "tablename", Value: tablename})
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
		logger.Info(&logger.Logs{Code: "UserData", Message: "No data found"})
		return nil, errors.New("no user found")
	}

	// Unmarshal it into actual task which front-end can understand as a JSON
	err = dynamodbattribute.UnmarshalMap(result.Items[0], user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal user record"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return user, nil
}

// FetchUser returns a single row of user depending on the username.
func FetchUser(ctx context.Context, username string, svc dynamodbiface.DynamoDBAPI) (*Response, error) {
	// Initialize a struct and returns a pointer to an instance of User struct
	var resp = new(Response)
	tablename := os.Getenv("USERS_TABLE")

	// Create the expression to fill the input struct
	// FilterExpression: "username = username_value"
	filter := expression.Name("username").Equal(expression.Value(username))

	// ProjectionExpression: id, first_name, last_name, username, email
	// SELECT id, first_name, last_name, username, email FROM tablename WHERE username = 'username_value'
	projection := expression.NamesList(expression.Name("id"), expression.Name("first_name"), expression.Name("last_name"), expression.Name("username"), expression.Name("email"))

	// Building expression with filter and projection to return the specific data and its columns
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

	// Gets the items and returns set of attributes for the item
	result, err := svc.Scan(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetItemError", Message: "Failed to get the user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "UserNotFound", Message: fmt.Sprintf("Could not find %s", username)})
		return nil, nil
	}

	// Unmarshal it into actual User struct which front-end can understand as a JSON
	err = dynamodbattribute.UnmarshalMap(result.Items[0], resp)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal user record"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return resp, nil
}

// UserAuthentication authenticates the user credentials.
func UserAuthentication(ctx context.Context, user *User, svc dynamodbiface.DynamoDBAPI) (*Response, error) {
	var resp = new(Response)
	tablename := os.Getenv("USERS_TABLE")
	validateAuth := user.ValidateAuhtentication()

	if validateAuth != "" {
		err := errors.New(validateAuth)

		logger.Error(err, &logger.Logs{Code: "UserAuthentication", Message: "Required fields are not entered"},
			logger.KVP{Key: "validation", Value: validateAuth})
		return nil, err
	}

	// Create the expression to fill the input struct
	// FilterExpression: "username = username_value AND password = password_value"
	filter := expression.Name("username").Equal(expression.Value(user.Username)).And(expression.Name("password").Equal(expression.Value(user.Password)))

	// ProjectionExpression: id, first_name, last_name, username, email
	projection := expression.NamesList(expression.Name("id"), expression.Name("first_name"), expression.Name("last_name"), expression.Name("username"), expression.Name("email"))

	// SELECT id, first_name, last_name, username, email WHERE username = 'username_value' AND password = 'password_value'
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

	// Make DynamoDB query API Call. Returns one or more items and item attributes by accessing them in a stable or secondary index
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
		return nil, errors.New("the username or password you entered is incorrect")
	}

	// Unmarshal a map into actual user which front-end can uderstand as a JSON
	err = dynamodbattribute.UnmarshalMap(result.Items[0], resp)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to unmarshal user record"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return resp, nil
}

// NewUserAccount creates a new object of user that will be saved on dynamoDB table.
func NewUserAccount(ctx context.Context, user *User, svc dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
	tablename := os.Getenv("USERS_TABLE")

	// Get the total cound of users to set UserID
	items := ItemCount(tablename, svc)
	user.UserID = fmt.Sprint(1 + items)

	// Converting the record to dynamodb.AttributeValue type
	value, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to marshal task"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Validate required fields
	validate := user.Validate()
	if validate != "" {
		err := errors.New(validate)

		logger.Error(err, &logger.Logs{Code: "UserAccountValidation", Message: "Required fields are not entered"},
			logger.KVP{Key: "validation", Value: validate})
		return nil, err
	}

	// Fetch the specific user account
	userExist, err := FetchUser(ctx, user.Username, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to fetch user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Checks if the user with the said username exist
	if userExist != nil {
		logger.Info(&logger.Logs{Code: "NewUserAccount", Message: "The username is already taken"})
		return nil, errors.New("username is already taken")
	}

	// Creating the data that you want to sent to dynamoDB
	params := &dynamodb.PutItemInput{
		// Map of attribute name-value pairs, one for each attribute
		Item:      value,
		TableName: aws.String(tablename),
	}

	// Creates a new item/object
	_, err = svc.PutItem(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to create a new user account"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return api.Response(http.StatusOK, user)
}

// UpdateUserAccount will update the specific user account information.
func UpdateUserAccount(ctx context.Context, user *User, svc dynamodbiface.DynamoDBAPI) (map[string]*dynamodb.AttributeValue, error) {
	tablename := os.Getenv("USERS_TABLE")

	// Get the specific user account
	userExist, err := GetUser(ctx, user.UserID, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to get user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Check if the task exist before updating
	if userExist == nil {
		logger.Info(&logger.Logs{Code: "UpdateUserData", Message: "The user does not exist"})
		return nil, errors.New("user does not exist")
	}

	// Validate if fields are empty
	user = user.ValidateEmpty(userExist)

	// Set a name-value pair for update expression
	update := expression.Set(expression.Name("first_name"), expression.Value(user.FirstName)).
		Set(expression.Name("last_name"), expression.Value(user.LastName)).
		Set(expression.Name("username"), expression.Value(user.Username)).
		Set(expression.Name("password"), expression.Value(user.Password))

	// Build the expression
	// UPDATE tablename SET first_name = first_name_value, last_name = last_name_value, ... WHERE id = user_id_value
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
			"id": {S: &user.UserID},
		},
	}

	// Updates the specific field
	output, err := svc.UpdateItemWithContext(ctx, params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to update user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return output.Attributes, nil
}

// DeleteAccount deletes a specific user based on the id given
func DeleteAccount(ctx context.Context, id string, svc dynamodbiface.DynamoDBAPI) (map[string]*dynamodb.AttributeValue, error) {
	tablename := os.Getenv("USERS_TABLE")

	// Get the specific user
	userExist, err := GetUser(ctx, id, svc)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to get user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	// Check if the use exist before deleting
	if userExist == nil {
		logger.Error(err, &logger.Logs{Code: "DeleteAccount", Message: "The user does not exist"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, errors.New("user does not exist")
	}

	// Create a delete query based on the parameter
	params := &dynamodb.DeleteItemInput{
		TableName:    aws.String(tablename),
		ReturnValues: aws.String("ALL_OLD"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	}

	// Deletes a single item in a table by primary key
	output, err := svc.DeleteItem(params)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DynamoDBError", Message: "Failed to delete user"},
			logger.KVP{Key: "tablename", Value: tablename})
		return nil, err
	}

	return output.Attributes, nil
}
