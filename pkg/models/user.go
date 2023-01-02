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
func UserResponse(result map[string]types.AttributeValue) (*Response, error) {
	user := new(Response)

	// Unmarshal it into actual task which front-end can understand as a JSON
	err := attributevalue.UnmarshalMap(result, user)
	if err != nil {
		return nil, err
	}

	return user, err
}

// GetUser gets a specific user using the id parameter.
func GetUser(ctx context.Context, id string) (*User, error) {
	// Initialze a struct and returns a pointer to an instance of User struct
	user := new(User)
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("USERS_TABLE")

	// Create the expression to fill the input struct
	// KeyConditionExpression: "id = user_id_value"
	builder.KeyExpression = expression.Key("id").Equal(expression.Value(id))

	// ProjectionExpression: id, first_name, last_name, username, email
	builder.Projection = expression.NamesList(expression.Name("id"), expression.Name("first_name"), expression.Name("last_name"), expression.Name("username"), expression.Name("email"), expression.Name("password"))

	// Make DynamoDB query API Call
	result, err := service.DynamoQuery(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetUser", Message: "Failed to execute DynamoQuery"}, logger.KVP{Key: "user_id", Value: id})
		return nil, err
	}

	// Checks if the result length is 0 to return a message of "No data found"
	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "GetUser", Message: "No data found"})
		return nil, errors.New("no user found")
	}

	// Unmarshal it into actual task which front-end can understand as a JSON
	err = attributevalue.UnmarshalMap(result.Items[0], user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GetUser", Message: "Failed to unmarshal user record"}, logger.KVP{Key: "user_id", Value: id})
		return nil, err
	}

	return user, nil
}

// GenerateUserID generates a new ID if the current ID argument passed already exist.
func GenerateUserID(ctx context.Context, id int) (string, error) {
	var users []User
	tablename := os.Getenv("USERS_TABLE")

	result, err := service.DynamoScan(ctx, tablename)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GenerateUserID", Message: "Failed to execute DynamoScan"})
		return "", err
	}

	// Unmarshal it into actual user for us to iterate
	err = attributevalue.UnmarshalListOfMaps(result.Items, &users)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "GenerateUserID", Message: "Failed to unmarshal user records"}, logger.KVP{Key: "user_id", Value: id})
		return "", err
	}

	for i := 0; i < len(users); i++ {
		userID, err := strconv.Atoi(users[i].UserID)
		if err != nil {
			logger.Error(err, &logger.Logs{Code: "GenerateUserID", Message: "Failed to convert user id string to integer"})
			return "", err
		}

		// Check if the current ID already exist
		if id == userID {
			// Loop backwards to check if the ID exist
			for j := (len(users) - 1); j >= 0; j-- {
				_userID, err := strconv.Atoi(users[j].UserID)
				if err != nil {
					logger.Error(err, &logger.Logs{Code: "GenerateUserID", Message: "Failed to convert user id string to integer"})
					return "", err
				}

				// Check backward if the current ID already exist
				if id == _userID {
					i = 0
					id -= 1
					continue
				}
			}
		}

	}

	return fmt.Sprint(id), nil
}

// FetchUser returns a single row of user depending on the username.
func FetchUser(ctx context.Context, username string) (*Response, error) {
	// Initialize a struct and returns a pointer to an instance of User struct
	var resp = new(Response)
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("USERS_TABLE")

	// Create the expression to fill the input struct
	// FilterExpression: "username = username_value"
	builder.Filter = expression.Name("username").Equal(expression.Value(username))

	// ProjectionExpression: id, first_name, last_name, username, email
	// SELECT id, first_name, last_name, username, email FROM tablename WHERE username = 'username_value'
	builder.Projection = expression.NamesList(expression.Name("id"), expression.Name("first_name"), expression.Name("last_name"), expression.Name("username"), expression.Name("email"))

	// Gets the items and returns set of attributes for the item
	result, err := service.DynamoScanExpression(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "FetchUser", Message: "Failed to execute DynamoScanExpression"})
		return nil, err
	}

	if len(result.Items) == 0 {
		logger.Info(&logger.Logs{Code: "FetchUser", Message: fmt.Sprintf("Could not find %s", username)})
		return nil, nil
	}

	// Unmarshal it into actual User struct which front-end can understand as a JSON
	err = attributevalue.UnmarshalMap(result.Items[0], resp)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "FetchUser", Message: "Failed to unmarshal user record"})
		return nil, err
	}

	return resp, nil
}

// UserAuthentication authenticates the user credentials.
func UserAuthentication(ctx context.Context, user *User) (*Response, error) {
	var resp = new(Response)
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("USERS_TABLE")
	validateAuth := user.ValidateAuhtentication()

	if validateAuth != "" {
		err := errors.New(validateAuth)
		logger.Error(err, &logger.Logs{Code: "UserAuthentication", Message: "Required fields are not entered"})

		return nil, err
	}

	// Create the expression to fill the input struct
	// FilterExpression: "username = username_value AND password = password_value"
	builder.Filter = expression.Name("username").Equal(expression.Value(user.Username)).And(expression.Name("password").Equal(expression.Value(user.Password)))

	// ProjectionExpression: id, first_name, last_name, username, email
	builder.Projection = expression.NamesList(expression.Name("id"), expression.Name("first_name"), expression.Name("last_name"), expression.Name("username"), expression.Name("email"))

	// Make DynamoDB query API Call. Returns one or more items.
	result, err := service.DynamoScanExpression(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "UserAuthentication", Message: "Failed to execute DynamoScanExpression"})
		return nil, err
	}

	// Checks if there are items returned
	if len(result.Items) == 0 {
		err := errors.New("the username or password you entered is incorrect")
		logger.Error(err, &logger.Logs{Code: "UserAuthentication", Message: "No data found"})

		return nil, err
	}

	// Unmarshal a map into actual user which front-end can uderstand as a JSON
	err = attributevalue.UnmarshalMap(result.Items[0], resp)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "UserAuthentication", Message: "Failed to unmarshal user record"})
		return nil, err
	}

	return resp, nil
}

// NewUserAccount creates a new object of user that will be saved on dynamoDB table.
func NewUserAccount(ctx context.Context, user *User) (*events.APIGatewayProxyResponse, error) {
	tablename := os.Getenv("USERS_TABLE")

	// Get the total cound of users to set UserID
	count := ItemCount(ctx, tablename) + 1

	// Checks if ID exist and generate a new one
	id, err := GenerateUserID(ctx, count)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "NewUserAccount", Message: "Failed to generate user id"})
		return api.StatusBadRequest(err)
	}
	user.UserID = id

	// Converting the record to dynamodb.AttributeValue type
	value, err := attributevalue.MarshalMap(user)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "NewUserAccount", Message: "Failed to marshal task"})
		return api.StatusBadRequest(err)
	}

	// Validate required fields
	validate := user.Validate()
	if validate != "" {
		err := errors.New(validate)
		logger.Error(err, &logger.Logs{Code: "NewUserAccount", Message: "Required fields are not entered"})

		return api.StatusBadRequest(err)
	}

	// Fetch the specific user account
	userExist, err := FetchUser(ctx, user.Username)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "NewUserAccount", Message: "Failed to fetch user"})
		return api.StatusBadRequest(err)
	}

	// Checks if the user with the said username exist
	if userExist != nil {
		err := errors.New("username is already taken")
		logger.Error(err, &logger.Logs{Code: "NewUserAccount", Message: "The username is already taken"})

		return api.StatusBadRequest(err)
	}

	// Creates a new item/object
	_, err = service.DynamoPutItem(ctx, tablename, value)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "NewUserAccount", Message: "Failed to execute DynamoPutItem"})
		return api.StatusBadRequest(err)
	}

	return api.Response(http.StatusOK, user)
}

// UpdateUserAccount will update the specific user account information.
func UpdateUserAccount(ctx context.Context, user *User) (map[string]types.AttributeValue, error) {
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("USERS_TABLE")

	// Get the specific user account
	userExist, err := GetUser(ctx, user.UserID)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "UpdateUserAccount", Message: "Failed to execute GetUser"})
		return nil, err
	}

	// Check if the task exist before updating
	if userExist == nil {
		err := errors.New("user does not exist")
		logger.Error(err, &logger.Logs{Code: "UpdateUserAccount", Message: "The user does not exist"})

		return nil, err
	}

	// Validate if fields are empty
	user = user.ValidateEmpty(userExist)

	builder.KeyAttribute = map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{
			Value: user.UserID,
		},
	}

	// Set a name-value pair for update expression
	builder.Update = expression.Set(expression.Name("first_name"), expression.Value(user.FirstName)).
		Set(expression.Name("last_name"), expression.Value(user.LastName)).
		Set(expression.Name("username"), expression.Value(user.Username)).
		Set(expression.Name("password"), expression.Value(user.Password))

	// Updates the specific field
	result, err := service.DynamoUpdateItem(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "UpdateUserAccount", Message: "Failed to execute DynamoUpdateItem"}, logger.KVP{Key: "user_id", Value: user.UserID})
		return nil, err
	}

	return result.Attributes, nil
}

// DeleteAccount deletes a specific user based on the id given.
func DeleteAccount(ctx context.Context, id string) (map[string]types.AttributeValue, error) {
	var builder service.DynamoBuilder
	builder.TableName = os.Getenv("USERS_TABLE")

	// Get the specific user
	userExist, err := GetUser(ctx, id)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteAccount", Message: "Failed to get user"})
		return nil, err
	}

	// Check if the use exist before deleting
	if userExist == nil {
		err := errors.New("user does not exist")
		logger.Error(err, &logger.Logs{Code: "DeleteAccount", Message: "The user does not exist"}, logger.KVP{Key: "user_id", Value: id})

		return nil, err
	}

	// Create a delete query based on the parameter
	builder.KeyAttribute = map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{
			Value: id,
		},
	}

	// Deletes a single item in a table by primary key
	result, err := service.DynamoDeleteItem(ctx, builder)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "DeleteAccount", Message: "Failed to delete user"}, logger.KVP{Key: "user_id", Value: id})
		return nil, err
	}

	return result.Attributes, nil
}
