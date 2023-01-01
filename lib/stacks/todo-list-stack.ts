// Import main CDK library as cdk
import * as cdk from 'aws-cdk-lib';
import { Stack, StackProps, RemovalPolicy } from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import { Runtime } from 'aws-cdk-lib/aws-lambda';
import { BucketEncryption, Bucket } from 'aws-cdk-lib/aws-s3';
import { BucketDeployment, Source } from 'aws-cdk-lib/aws-s3-deployment';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import { BillingMode } from 'aws-cdk-lib/aws-dynamodb';
import * as apigw from 'aws-cdk-lib/aws-apigateway';
import * as iam from 'aws-cdk-lib/aws-iam';
import { ManagedPolicy } from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class TodoListAppStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const REMOVAL_POLICY = RemovalPolicy.DESTROY; // when the resource is removed from the app, it will be physically destroyed

    // Defining an IAM Role
    const TodoList_customRole = new iam.Role(this, 'TodoList_customRole', {
      roleName: 'TodoList_customRole',                                                      // name for IAM role
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),                          // represents an AWS service
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole")  // managed policies associated with this role
      ]
    });
    TodoList_customRole.applyRemovalPolicy(REMOVAL_POLICY);

    // S3 Bucket for our Frontend
    // Instantiate Bucket with bucketName and versioned properties
    const bucket = new Bucket(this, 'TodoListFrontEnd', {
      bucketName: 'todo-list-app-dev',          // physical name of the bucket
      websiteIndexDocument: 'index.html',       // enable static web hosting
      websiteErrorDocument: 'error.html',       // error document (404.html)
      publicReadAccess: true,                   // public read access on the bucket
      removalPolicy: REMOVAL_POLICY,
      autoDeleteObjects: true,                  // allow these files to be removed along with the bucket
      encryption: BucketEncryption.S3_MANAGED,  // server-side encryption with a master key managed by S3
    });

    // Deploying S3 Bucket static web application
    const webApplication = new BucketDeployment(this, 'WebTodoListApplication', {
      sources: [ Source.asset(`web_app/`) ],            // sources from which to deploy the contents of this bucket
      destinationBucket: bucket,                        // S3 bucket to sync the contents of the zip file to
    });
    
    // DynamoDB will serve as our database to store our data
    const todoListUsersTable = new dynamodb.Table(this, 'TodoListUsersTable', {
      tableName: 'todo-list-users',             // physical table name
      billingMode: BillingMode.PAY_PER_REQUEST, // pay for what you use
      partitionKey: {                           // partition key attribute definition
        name: 'id',
        type: dynamodb.AttributeType.STRING,    // data type for the attribute
      },
      removalPolicy: REMOVAL_POLICY,            
    });

    const todoListTaskTable = new dynamodb.Table(this, 'TodoListTasksTable', {
      tableName: 'todo-list-tasks',              // physical table name
      billingMode: BillingMode.PAY_PER_REQUEST,  // pay for what you use
      partitionKey: {                            // partition key attribute definition
        name: 'id',
        type: dynamodb.AttributeType.STRING      // data type for the attribute
      },
      removalPolicy: REMOVAL_POLICY
    });

    // Lambda Function
    const todoList = new lambda.Function(this, 'TodoListFunction', {
      runtime: Runtime.GO_1_X,                                      // runtime environment for the Lambda function that you are uploading
      handler: 'todoList',                                          // name of the method within your code that Lambda calls to execute your function
      code: lambda.Code.fromAsset('cmd/todoList'),                  // source code of your Lambda function; code loaded from "lambda" directory
      functionName: 'TodoListApplication',                          // name for the function
      description: 'Triggered when there is an API request',        // description of the function
      timeout: cdk.Duration.seconds(60),                            // a time for how long will the lambda run
      memorySize: 1024,                                             // amount of memory in MB that is allocated to your Lambda function
      environment: {                                                // key-value pairs that Lambda caches and makes available for your Lambda functions
        "USERS_TABLE": todoListUsersTable.tableName,
        "TASKS_TABLE": todoListTaskTable.tableName
      },
    });

    // Granting Lambda function to read and write to DynamoDB Tables
    todoListUsersTable.grantReadWriteData(todoList);
    todoListTaskTable.grantReadWriteData(todoList);
    

    // API Gateway : All routes (and REST verbs) will pass through the Lambda function
    const api = new apigw.LambdaRestApi(this, 'todo-list-api', {
      proxy: false,                                                                         // route all requests to the Lambda function
      restApiName: 'todo-list-api',                                                         // name for the API Gateway RestApi resource
      handler: todoList,                                                                    // default Lambda function that handles all requests from this API
      description: 'Holds all the requests coming from the API for Todo List Application',  // description of the purpose of this API Gateway RestApi resource
      apiKeySourceType: apigw.ApiKeySourceType.HEADER,
      defaultMethodOptions: {
        apiKeyRequired: true,
        
      },
      defaultCorsPreflightOptions: {
        statusCode: 200,
        allowOrigins: apigw.Cors.ALL_ORIGINS,
        allowMethods: ['GET', 'POST', 'DELETE', 'OPTIONS'],
        allowHeaders: ['Content-Type', 'X-Api-Key', 'user_id', 'task_id'],
      }
    });
    api.applyRemovalPolicy(REMOVAL_POLICY);           // controls what happens to this resource when it stops being managed by CloudFormation

    const plan = api.addUsagePlan('todo-list-usage-plan', {
      name: 'todo-list-usage-plan'
    });

    const key = api.addApiKey('todo-list-api-key', {
      apiKeyName: 'todo-list-api-key',
      value: 'VGhpcyBpcyBhIFZlcnkgRWFzeSBQYXNzd29yZCAyMDAl',
    });

    plan.addApiKey(key);
    plan.addApiStage({ stage: api.deploymentStage });
    plan.applyRemovalPolicy(REMOVAL_POLICY);

    const tasks = api.root.addResource('tasks');      // represents the root resource of the API endpoint
    tasks.addMethod('GET');                           // GET /tasks
    tasks.addMethod('POST');
    tasks.addMethod('DELETE');

    const task = tasks.addResource('{task_id}');
    task.addMethod('GET');                            // GET /tasks/{task_id}
    task.addMethod('POST');                           // POST /tasks/{task_id}
    task.addMethod('DELETE');                         // DELETE /tasks/{task_id}

    const users = api.root.addResource('users');
    users.addMethod('GET');                           // GET /users
    users.addMethod('POST');

    const user = users.addResource('{user_id}');
    user.addMethod('POST');                           // POST /users/{user_id}
    user.addMethod('DELETE');                         // DELETE /users/{user_id}

    const ResponseParameters = {
      'method.response.header.Content-Type': true,
      'method.response.header.Access-Control-Allow-Origin': true,
      'method.response.header.Access-Control-Allow-Headers': true
    }

    // Defining the JSON Schema for the transformed valid response    
    const TodoListApiResponseModel = api.addModel('TodoListApiResponseModel', {
      modelName: 'TodoListApiResponseModel',
      schema: {
        schema: apigw.JsonSchemaVersion.DRAFT4,
        type: apigw.JsonSchemaType.OBJECT
      }
    });

    // Defining the JSON Schema for the transformed error response
    const TodoListApiErrorResponseModel = api.addModel('TodoListApiErrorResponseModel', {
      modelName: 'TodoListApiErrorResponseModel',
      schema: {
        schema: apigw.JsonSchemaVersion.DRAFT4,
        type: apigw.JsonSchemaType.OBJECT
      }
    });

    // Adding Method Response for StatusOK
    api.methods.forEach((method) => method.addMethodResponse({
      statusCode: "200",
      responseParameters: ResponseParameters,
      responseModels: { 'application/json': TodoListApiResponseModel }
    }));

    // Adding Method Response for StatusBadRequest
    api.methods.forEach((method) => method.addMethodResponse({
      statusCode: "400",
      responseParameters: ResponseParameters,
      responseModels: { 'application/json': TodoListApiErrorResponseModel }
    }));

    api.methods.filter((method) => method.httpMethod === "OPTIONS")
               .forEach((method) => {
                const methodCfn = method.node.defaultChild as apigw.CfnMethod;
                methodCfn.addPropertyOverride('ApiKeyRequired', false);
                methodCfn.addPropertyDeletionOverride('AuthorizedId');
                methodCfn.addPropertyOverride('AuthorizationType', 'NONE');
               });
  }
}