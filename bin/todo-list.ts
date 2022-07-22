#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { TodoListAppStack } from '../lib/stacks/todo-list-stack';

const app = new cdk.App();
new TodoListAppStack(app, 'TodoListAppStack', {
   env: {
      account: process.env.CDK_DEFAULT_ACCOUNT,
      region: process.env.CDK_DEFAULT_REGION
   }
});