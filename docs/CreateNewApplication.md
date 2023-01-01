# Creating new application

To create a new application, you'll need the following:
* IDE
* AWS CLI
* Node.js
* AWS Account and User
* Your preferred programming language

If you're not still familiar with AWS CDK, you can go through the AWS CDK Workshop for an introduction to [AWS CDK TypeScript](https://cdkworkshop.com/20-typescript.html). You could also visit [AWS CDK Reference Documentation v2](https://docs.aws.amazon.com/cdk/api/v2/).

To create a new AWS CDK TypeScript project, you should create an empty directory and initialize it with a language of `TypeScript`.

```bash
dev@dev:$ mkdir app-name && cd app-name
dev@dev:~:app-name$ cdk init app --language typescript
Applying project template app for typescript
# Welcome to your CDK TypeScript project

This is a blank project for TypeScript development with CDK.

The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

* `npm run build`   compile typescript to js
* `npm run watch`   watch for changes and compile
* `npm run test`    perform the jest unit tests
* `cdk deploy`      deploy this stack to your default AWS account/region
* `cdk diff`        compare deployed stack with current state
* `cdk synth`       emits the synthesized CloudFormation template

Initializing a new git repository...
Executing npm install...
npm WARN deprecated source-map-url@0.4.1: See https://github.com/lydell/source-map-url#deprecated
npm WARN deprecated urix@0.1.0: Please see https://github.com/lydell/urix#deprecated
npm WARN deprecated resolve-url@0.2.1: https://github.com/lydell/resolve-url#deprecated
npm WARN deprecated source-map-resolve@0.5.3: See https://github.com/lydell/source-map-resolve#deprecated
npm WARN deprecated sane@4.1.0: some dependency vulnerabilities fixed, support for node < 10 dropped, and newer ECMAScript syntax/features added
✅ All done!
****************************************************
*** Newer version of CDK is available [2.32.1]   ***
*** Upgrade recommended (npm install -g aws-cdk) ***
****************************************************
```

After running the `cdk init` command, you'll see a bunch of files created on your directory. The `bin/app-name.ts` will be your app's entry point while the `lib/app-name-stack.ts` will be your main stack.

To synthesize your CDK application, you can use the `cdk synth` command.
```bash
dev@dev:~:app-name$ cdk synth
Resources:
  CDKMetadata:
    Type: AWS::CDK::Metadata
    Properties:
      Analytics: v2:deflate64:H4sIAAAAAAAA/zPSM7TQM1BMLC/WTU7J1s3JTNKrDi5JTM7WcU7LC0otzi8tSk4FsZ3z81IySzLz82p18vJTUvWyivXLDM30DE30jBSzijMzdYtK80oyc1P1giA0AM07vtBZAAAA
    Metadata:
      aws:cdk:path: AppNameStack/CDKMetadata/Default
    Condition: CDKMetadataAvailable
Conditions:
  CDKMetadataAvailable:
    Fn::Or:
      - Fn::Or:
          - Fn::Equals:
              - Ref: AWS::Region
              - af-south-1
          - Fn::Equals:
              - Ref: AWS::Region
              - ap-east-1
          - Fn::Equals:
              - Ref: AWS::Region
              - ap-northeast-1
          - Fn::Equals:
              - Ref: AWS::Region
              - ap-northeast-2
          - Fn::Equals:
              - Ref: AWS::Region
              - ap-south-1
          - Fn::Equals:
              - Ref: AWS::Region
              - ap-southeast-1
          - Fn::Equals:
              - Ref: AWS::Region
              - ap-southeast-2
          - Fn::Equals:
              - Ref: AWS::Region
              - ca-central-1
          - Fn::Equals:
              - Ref: AWS::Region
              - cn-north-1
          - Fn::Equals:
              - Ref: AWS::Region
              - cn-northwest-1
      - Fn::Or:
          - Fn::Equals:
              - Ref: AWS::Region
              - eu-central-1
          - Fn::Equals:
              - Ref: AWS::Region
              - eu-north-1
          - Fn::Equals:
              - Ref: AWS::Region
              - eu-south-1
          - Fn::Equals:
              - Ref: AWS::Region
              - eu-west-1
          - Fn::Equals:
              - Ref: AWS::Region
              - eu-west-2
          - Fn::Equals:
              - Ref: AWS::Region
              - eu-west-3
          - Fn::Equals:
              - Ref: AWS::Region
              - me-south-1
          - Fn::Equals:
              - Ref: AWS::Region
              - sa-east-1
          - Fn::Equals:
              - Ref: AWS::Region
              - us-east-1
          - Fn::Equals:
              - Ref: AWS::Region
              - us-east-2
      - Fn::Or:
          - Fn::Equals:
              - Ref: AWS::Region
              - us-west-1
          - Fn::Equals:
              - Ref: AWS::Region
              - us-west-2
Parameters:
  BootstrapVersion:
    Type: AWS::SSM::Parameter::Value<String>
    Default: /cdk-bootstrap/hae659cws/version
    Description: Version of the CDK Bootstrap resources in this environment, automatically retrieved from SSM Parameter Store. [cdk:skip]
Rules:
  CheckBootstrapVersion:
    Assertions:
      - Assert:
          Fn::Not:
            - Fn::Contains:
                - - "1"
                  - "2"
                  - "3"
                  - "4"
                  - "5"
                - Ref: BootstrapVersion
        AssertDescription: CDK bootstrap stack version 6 required. Please run 'cdk bootstrap' with a recent version of the CDK CLI.
```

Before you deploy your AWS CDK application into an environment, you can install the bootstrap stack using the `cdk bootstrap` command.
```bash
dev@dev:~:app-name$ cdk bootstrap
 ⏳  Bootstrapping environment aws://xxxxxxxxxx/xx-xxxx-x...
Trusted accounts for deployment: (none)
Trusted accounts for lookup: (none)
Using default execution policy of 'arn:aws:iam::aws:policy/AdministratorAccess'. Pass '--cloudformation-execution-policies' to customize.
CDKToolkit: creating CloudFormation changeset...

 ✅  Environment aws://xxxxxxxxxx/xx-xxxx-x bootstrapped.
```

After bootstrapping your application, you can now deploy it using `cdk deploy` command.
```bash
dev@dev:~:app-name$ cdk deploy
This deployment will make potentially sensitive changes according to your current security approval level (--require-approval broadening).
Please confirm you intend to make the following modifications:

IAM Statement Changes
┌───┬────────────────────────────────┬────────┬─────────────────┬────────────────────────────────┬────────────────────────────────┐
│   │ Resource                       │ Effect │ Action          │ Principal                      │ Condition                      │
├───┼────────────────────────────────┼────────┼─────────────────┼────────────────────────────────┼────────────────────────────────┤
│ + │ ${CdkWorkshopQueue.Arn}        │ Allow  │ sqs:SendMessage │ Service:sns.amazonaws.com      │ "ArnEquals": {                 │
│   │                                │        │                 │                                │   "aws:SourceArn": "${CdkWorks │
│   │                                │        │                 │                                │ hopTopic}"                     │
│   │                                │        │                 │                                │ }                              │
└───┴────────────────────────────────┴────────┴─────────────────┴────────────────────────────────┴────────────────────────────────┘
(NOTE: There may be security-related changes not in this list. See https://github.com/aws/aws-cdk/issues/1299)

Do you wish to deploy these changes (y/n)?
```

This is warning you that deploying the app contains security-sensitive changes. Since we need to allow the topic to send messages to the queue, enter `y` to deploy the stack and create the resources. If you've successfully deployed your stack resources, you can use the **AWS CloudFormation console** to manage your stacks.

## Remove CDK Bootstrap
```bash
dev@dev:~:app-name$ cdk bootstrap --destroy
dev@dev:~:app-name$ cdk bootstrap --clean
```

It is also important to delete your stack at the **CloudFormation** service or you can remove it using CLI. The `rb` command removes the bucket. Adding `--force` parameter will first remove all the objects in the bucket and then remove the bucket itself.

```bash
dev@dev:~:app-name$ aws cloudformation delete-stack --stack-name CDKToolkit
dev@dev:~:app-name$ aws s3 ls | grep cdktoolkit                            # copy the name
dev@dev:~:app-name$ aws s3 rb --force s3://cdktoolkit-stagingbucket-abcdef # replace the name here
```

## Compile changes made to TypeScript file automatically
It will compile any `.ts` file.
```bash
dev@dev:~:app-name$ tsc -w
```

## Building your function (if you are using GO)
Preparing a binary to deploy to AWS Lambda requires that it is compiled for Linux and placed into a .zip file.

#### Linux and MacOS
```bash
dev@dev:~:app-name$ GOOS=linux GOARCH=amd64 go build -o main main.go
dev@dev:~:app-name$ zip main.zip main
```