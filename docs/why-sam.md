# Why Serverless Application Model (SAM)
### For Go Functions-as-a-Service

We see how Functions-as-a-Service invites new techniques to [develop, package and deploy](dev-package-deploy.md) our apps versus monolithic apps. This poses a question: what tools or frameworks make FaaS configuration, development and deployment easy so we can focus on our code?

AWS offers a first-party and [open-source](https://aws.amazon.com/about-aws/whats-new/2018/04/aws-sam-implementation-is-now-open-source/) option, the Serverless Application Model (SAM). [The docs](https://docs.aws.amazon.com/lambda/latest/dg/serverless_app.html) explain:

> The AWS Serverless Application Model (AWS SAM) is a model to define serverless applications. AWS SAM is natively supported by AWS CloudFormation and defines simplified syntax for expressing serverless resources. The specification currently covers APIs, Lambda functions and Amazon DynamoDB tables. SAM is available under Apache 2.0 for AWS partners and customers to adopt and extend within their own toolsets.

This is one of many approaches for building FaaS apps. There are other open-source frameworks that target Lambda:

- [Apex Up](https://up.docs.apex.sh/)
- [Serverless Framework](https://serverless.com/)
- [Terraform](https://www.terraform.io/docs/providers/aws/r/lambda_function.html)

And other AWS tools and strategies such as:

- [Blueprints](https://docs.aws.amazon.com/lambda/latest/dg/get-started-create-function.html)
- [Chalice](https://github.com/aws/chalice)
- [Cloud9 IDE](https://docs.aws.amazon.com/cloud9/latest/user-guide/lambda-functions.html)
- [CloudFormation](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html)

So why SAM over other options?

## SAM - CloudFormation Simplified

There are clear best practices for building systems on AWS. One is **Infrastructure-as-Code**. This is a strategy where we declare all the desired infrastructure in a configuration file and pass this config to a system that can:

* Create infrastructure from scratch
* Show the steps to migrate existing infrastructure to a new config
* Execute changes to migrate to a new config
* Destroy all the infrastructure

While it is much easier to create our first Lambda functions by clicking around in the AWS UI, infrastructure-as-code offers development best practices like code reviews for infrastructure changes, and operational best practices like change plan reviews and rollback strategies.

The primary choices for infrastructure-as-code are CloudFormation and Terraform. As long as we are targeting AWS only, it's hard to beat the utility of the CloudFormation service. 

First, CloudFormation offers a way define virtually any AWS architecture with a JSON or YAML template. It supports 100s of AWS resource types, from API Gateways to RDS databases, and lets us customize these resources through parameters for domain names, amount of CPU, memory or storage, etc.

Next, CloudFormation is a managed AWS service -- a set of APIs -- that automate everything around our infrastructure. We POST a config file template to the CloudFormation API and AWS will automatically create all the resources in a matter of minutes, with more APIs to describe progress along the way. Then we present a new version of our template to the API and AWS will tell us what it plans to change, then safely execute the change with little to no downtime. Finally the APIs offer a strategy to roll back changes or completely destroy all the resources. With CloudFormation we don't have to worry about managing the state of existing infrastructure or talking to individual AWS service APIs to create or update resources.

There's one big catch with CloudFormation: writing templates can be hard in general, and particularly difficult for medium to large sets of infrastructure.

SAM vastly simplifies the config file format.

SAM is a dialect of CloudFormation that focuses on key FaaS services like API Gateway, Lambda functions and DynamoDB tables. It offers us a simplified YAML specification for defining functions, events and permissions. 

```yaml
---
AWSTemplateFormatVersion: '2010-09-09'

Resources:
  DashboardFunction:
    Properties:
      CodeUri: ./handlers/dashboard
      Events:
        Request:
          Properties:
            Method: GET
            Path: /
          Type: Api
      Handler: main
      Runtime: go1.x
    Type: AWS::Serverless::Function

Transform: AWS::Serverless-2016-10-31
```
> A few lines of config creates a Lambda function, API Gateway and route

Behind the scenes SAM transforms the FaaS resource config into standard CloudFormation resources. SAM also passes through standard CloudFormation resources untouched, so we retain support for the 100s of other AWS resource types.

SAM is the infrastructure-as-code pattern made simple and more compact for FaaS apps, but doesn't restrict us from following the infrastructure-as-code best practice for all of our AWS resources.

## SAM - AWS Constrained

Another best practice for building systems on AWS is **minimal infrastructure**.

AWS is an infinitely capable platform, so there are countless ways to build a system. We could build an app with a classic ELB, EC2, EBS services and our own userdata scripts that install nginx, Redis, and Postgres for handling HTTP requests, queuing, storage and persistance respectively. However here we are signing up for a lot of work configuring and managing servers, storage and software.

As AWS and the IaaS ecosystem evolves, we are all are steering away from raw infrastructure (e.g. EC2 + EBS) and towards high level managed cloud services (e.g. RDS).

Hence the "serverless" approach of API Gateway, Lambda, SQS, DynamoDB and S3 for HTTP, compute, queuing, data and storage respectively. These services are building blocks for our app and infrastructure we are no longer responsible for.

SAM focuses on [three primary resources](https://github.com/awslabs/serverless-application-model/blob/develop/versions/2016-10-31.md#resource-types) -- APIs, functions and tables -- that form the skeleton of our app. It then offers [10 types of events](https://github.com/awslabs/serverless-application-model/blob/develop/versions/2016-10-31.md#event-source-types) -- e.g. API Gateway, S3, SNS, DynamoDB, CloudWatch scheduled events -- to connect our functions together.

These components allow us to build very capable apps on extremely minimal infrastructure.

With minimal infrastructure we spend less time configuring and operating infrastructure, and more time writing our business code.

## AWS Config - Hello World

Let's look more closely at a SAM template for a simple [http function](http-functions.md) with access to an S3 bucket:

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
> From [template.yml](../template.yml)

Here we've configured Lambda, API gateway, S3 and per-function permissions in a few lines of YAML.

With SAM we can launch this app locally with the `sam local start-api` command and develop our code. Then we can deploy it with the `sam package` and `sam deploy` commands.

Behind the scenes SAM turns this into a CloudFormation template. It is demonstrative how complicated this is compared to source the SAM template.

<details>
<summary>Full CloudFormation template JSON...</summary>

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
</details>
&nbsp;

Thats a lot of config to describe our FaaS infrastructure, but SAM makes writing the config, and development, packaging and deployment simple.

## Summary

Building an app with SAM guides us towards AWS best practices for:

- Infrastructure-as-code
- Minimal infrastructure

We no longer have to worry about:

- Configuring AWS manually
- Writing complex CloudFormation templates
- Using 3rd party middleware or frameworks

SAM makes it very easy to build and maintain a FaaS app.