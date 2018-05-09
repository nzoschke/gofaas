# Why Serverless Application Model (SAM)
### For Go Functions-as-a-Service

We have seen how Functions-as-a-Service invites new techniques to [develop, package and deploy](docs/dev-package-deploy) our apps versus monolithic apps. This poses a quest: what tools or frameworks make FaaS development, configuration and deployment easy so we can focus on our code?

AWS offers a first-party option -- the [Serverless Application Model](https://docs.aws.amazon.com/lambda/latest/dg/serverless_app.html):

> The AWS Serverless Application Model (AWS SAM) is a model to define serverless applications. AWS SAM is natively supported by AWS CloudFormation and defines simplified syntax for expressing serverless resources. The specification currently covers APIs, Lambda functions and Amazon DynamoDB tables. SAM is available under Apache 2.0 for AWS partners and customers to adopt and extend within their own toolsets.

This is one of many approaches for building FaaS apps. There are other open-source frameworks that target Lambda:

- [Apex Up](https://up.docs.apex.sh/)
- [Serverless Framework](https://serverless.com/)
- [Terraform](https://www.terraform.io/docs/providers/aws/r/lambda_function.html)

AWS offers other strategies such as:

- [Blueprints](https://docs.aws.amazon.com/lambda/latest/dg/get-started-create-function.html)
- [Chalice](https://github.com/aws/chalice)
- [Cloud9 IDE](https://docs.aws.amazon.com/cloud9/latest/user-guide/lambda-functions.html)
- [AWS CloudFormation](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html)

So why SAM over other options? 

## SAM - CloudFormation Simplified

There are clear best practices for building systems on AWS. One is **Infrastructure-as-Code**.

While it's very easy to create one Lambda function manually by clicking and coding in the AWS UI, it can be very hard to maintain this over time. Instead we prefer to use a codebase tracked in revision control and describe AWS infrastructure with configuration files, such as AWS CloudFormation templates or Terraform config files. This lets us evolve the AWS config over time with development best practices like code reviews, and ops best practices like infrastructure change plan reviews and rollbacks.

As long as you are targeting AWS only, it's hard to beat the utility of the CloudFormation service for infrastructure-as-code.

First, CloudFormation offers a way define virtually any AWS architecture with a JSON or YAML template, including parameters for things we may customize like a domain name, amount of CPU, memory or storage for a resource.

Next, CloudFormation is a managed AWS service -- a set of APIs that automate everything about our infrastructure. Effectively we present a template to the CloudFormation API and AWS will safely create all the resources. Then we present a new version of our template to the API and AWS will tell us what it plans to change, then safely executes the change with zero downtime. Finally the APIs offer a strategy to roll back changes or completely destroy all the resources. With CloudFormation we don't have to worry about managing the state of a stack or talking to individual AWS service APIs to create or update resources.

There's one big catch with CloudFormation: writing templates can be challenging for medium to large stacks.

SAM fixes this.

SAM is a dialect of CloudFormation that focuses on key FaaS services like API Gateway, Lambda functions and DynamoDB tables. It offers us a vastly simplified YAML specification for defining functions, events and permissions. Behind the scenes it transforms this configuration file into a CloudFormation configuration file.

SAM is the infrastructure-as-code pattern vastly simplified for FaaS apps. It is still CloudFormation so we don't need to depend on any 3rd party tools or frameworks to manage our functions.

## SAM - AWS Constrained

Another best practice for building systems on AWS is **minimal infrastructure**.

AWS is an infinitely capable platform, so there are countless ways to build a particular system. We could build our gofaas system with the classic ELB, EC2, EBS services and our own install of nginx, Redis, and Postgres for handling HTTP requests, queuing, storage and persistance. However here we are signing up for a lot of work on managing servers, storage and software.

But as AWS and the IaaS ecosystem evolves, we AWS and us users are steering towards managed cloud services.

We are now reaching towards API Gateway, Lambda, SQS, DynamoDB and S3 for HTTP, compute, queuing, data and storage respective. These services are building blocks for our app and infrastructure we are no longer responsible for.

SAM focuses on [three primary resources](https://github.com/awslabs/serverless-application-model/blob/develop/versions/2016-10-31.md#resource-types) -- APIs, functions and tables -- that form the skeleton of our app. It then offers [10 types of events](https://github.com/awslabs/serverless-application-model/blob/develop/versions/2016-10-31.md#event-source-types) -- e.g. S3, SNS, DynamoDB, CloudWatch Schedule -- that represent the services that our functions interact with.

These components allow us to build very capable apps on extremely minimal infrastructure.

The result is less time configuring and operating infrastructure, and more time writing our business code.

## AWS Config - Hello World

Let's look back at a SAM template for a simple [http function](docs/http-functions.md) with access to an S3 bucket:

```yaml
---
AWSTemplateFormatVersion: '2010-09-09'

Resources:
  Bucket:
    Type: AWS::S3::Bucket

  DashboardFunction:
    Properties:
      CodeUri: ./handlers/dashboard/main.zip
      Environment:
        Variables:
          BUCKET: !Ref Bucket
      Events:
        Request:
          Properties:
            Method: GET
            Path: /
          Type: Api
      Handler: main
      Policies:
        - S3CrudPolicy:
            BucketName: !Ref Bucket
      Runtime: go1.x
    Type: AWS::Serverless::Function

Transform: AWS::Serverless-2016-10-31
```
> From [template.yml](template.yml)

Here we've configured Lambda and API gateway in a few lines of YAML.

With SAM we can launch this app locally with the `sam local start-api` command, and we can deploy it with the `sam package` and `sam deploy` command.

Behind the scenes SAM turns this into a CloudFormation template. It is demonstrative how complicated this is compared to the SAM template.

```json
{
    "AWSTemplateFormatVersion": "2010-09-09",
    "Resources": {
        "DashboardFunctionRequestPermissionTest": {
            "Type": "AWS::Lambda::Permission",
            "Properties": {
                "Action": "lambda:invokeFunction",
                "Principal": "apigateway.amazonaws.com",
                "FunctionName": {
                    "Ref": "DashboardFunction"
                },
                "SourceArn": {
                    "Fn::Sub": [
                        "arn:aws:execute-api:${AWS::Region}:${AWS::AccountId}:${__ApiId__}/${__Stage__}/GET/",
                        {
                            "__Stage__": "*",
                            "__ApiId__": {
                                "Ref": "ServerlessRestApi"
                            }
                        }
                    ]
                }
            }
        },
        "ServerlessRestApiDeployment207a898e93": {
            "Type": "AWS::ApiGateway::Deployment",
            "Properties": {
                "RestApiId": {
                    "Ref": "ServerlessRestApi"
                },
                "Description": "RestApi deployment id: 207a898e93177ae2f10ef6b27049781b90451f34",
                "StageName": "Stage"
            }
        },
        "DashboardFunction": {
            "Type": "AWS::Lambda::Function",
            "Properties": {
                "Code": {
                    "S3Bucket": "pkgs-572007530218-us-east-1",
                    "S3Key": "8e2a35b75c326919518e0471f34b3c02"
                },
                "Tags": [
                    {
                        "Value": "SAM",
                        "Key": "lambda:createdBy"
                    }
                ],
                "Environment": {
                    "Variables": {
                        "BUCKET": {
                            "Ref": "Bucket"
                        }
                    }
                },
                "Handler": "main",
                "Role": {
                    "Fn::GetAtt": [
                        "DashboardFunctionRole",
                        "Arn"
                    ]
                },
                "Runtime": "go1.x"
            }
        },
        "ServerlessRestApiProdStage": {
            "Type": "AWS::ApiGateway::Stage",
            "Properties": {
                "DeploymentId": {
                    "Ref": "ServerlessRestApiDeployment207a898e93"
                },
                "RestApiId": {
                    "Ref": "ServerlessRestApi"
                },
                "StageName": "Prod"
            }
        },
        "DashboardFunctionRole": {
            "Type": "AWS::IAM::Role",
            "Properties": {
                "ManagedPolicyArns": [
                    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
                ],
                "Policies": [
                    {
                        "PolicyName": "DashboardFunctionRolePolicy0",
                        "PolicyDocument": {
                            "Statement": [
                                {
                                    "Action": [
                                        "s3:GetObject",
                                        "s3:ListBucket",
                                        "s3:GetBucketLocation",
                                        "s3:GetObjectVersion",
                                        "s3:PutObject",
                                        "s3:GetLifecycleConfiguration",
                                        "s3:PutLifecycleConfiguration"
                                    ],
                                    "Resource": [
                                        {
                                            "Fn::Sub": [
                                                "arn:${AWS::Partition}:s3:::${bucketName}",
                                                {
                                                    "bucketName": {
                                                        "Ref": "Bucket"
                                                    }
                                                }
                                            ]
                                        },
                                        {
                                            "Fn::Sub": [
                                                "arn:${AWS::Partition}:s3:::${bucketName}/*",
                                                {
                                                    "bucketName": {
                                                        "Ref": "Bucket"
                                                    }
                                                }
                                            ]
                                        }
                                    ],
                                    "Effect": "Allow"
                                }
                            ]
                        }
                    }
                ],
                "AssumeRolePolicyDocument": {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Action": [
                                "sts:AssumeRole"
                            ],
                            "Effect": "Allow",
                            "Principal": {
                                "Service": [
                                    "lambda.amazonaws.com"
                                ]
                            }
                        }
                    ]
                }
            }
        },
        "DashboardFunctionRequestPermissionProd": {
            "Type": "AWS::Lambda::Permission",
            "Properties": {
                "Action": "lambda:invokeFunction",
                "Principal": "apigateway.amazonaws.com",
                "FunctionName": {
                    "Ref": "DashboardFunction"
                },
                "SourceArn": {
                    "Fn::Sub": [
                        "arn:aws:execute-api:${AWS::Region}:${AWS::AccountId}:${__ApiId__}/${__Stage__}/GET/",
                        {
                            "__Stage__": "Prod",
                            "__ApiId__": {
                                "Ref": "ServerlessRestApi"
                            }
                        }
                    ]
                }
            }
        },
        "Bucket": {
            "Type": "AWS::S3::Bucket"
        },
        "ServerlessRestApi": {
            "Type": "AWS::ApiGateway::RestApi",
            "Properties": {
                "Body": {
                    "info": {
                        "version": "1.0",
                        "title": {
                            "Ref": "AWS::StackName"
                        }
                    },
                    "paths": {
                        "/": {
                            "get": {
                                "x-amazon-apigateway-integration": {
                                    "httpMethod": "POST",
                                    "type": "aws_proxy",
                                    "uri": {
                                        "Fn::Sub": "arn:aws:apigateway:${AWS::Region}:lambda:path/2015-03-31/functions/${DashboardFunction.Arn}/invocations"
                                    }
                                },
                                "responses": {}
                            }
                        }
                    },
                    "swagger": "2.0"
                }
            }
        }
    }
}
```

## Summary

Building an app with SAM guides us towards best practices for:

- Infrastructure-as-code
- Minimal infrastructure

We no longer have to worry about:

- Complex CloudFormation templates
- 3rd party middleware or frameworks

Our SAM app is very easy to build and maintain.