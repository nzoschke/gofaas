# Per-Function Policies
### With Lambda, IAM and SAM

Building an application as a constellation of functions offers a huge security advantage. Every function can have its own set of environment variables and its own policy over what other resource APIs it has permission to use.

## AWS Config

Consider our worker and user functions.

The user functions have policies that grants them permission to manage records in a single DynamoDB table via the `DynamoDBCrudPolicy`. The create function has permission to encrypt data, and the read function has permission to decrypt data with KMS via custom statements. Both have `KEY_ID` and `TABLE_NAME` environment variables so the functions know what AWS resources to use.

```yaml
Resources:
  Key:
    Properties: ...
    Type: AWS::KMS::Key

  UsersTable:
    Properties: ...
    Type: AWS::Serverless::SimpleTable

  UserCreateFunction:
    Properties:
      Environment:
        Variables:
          KEY_ID: !Ref Key
          TABLE_NAME: !Ref UsersTable
      Policies:
        - DynamoDBCrudPolicy:
            TableName: !Ref UsersTable
        - Statement:
            - Action:
                - kms:Encrypt
              Effect: Allow
              Resource: !GetAtt Key.Arn
          Version: 2012-10-17
    Type: AWS::Serverless::Function

  UserReadFunction:
    Properties:
      Environment:
        Variables:
          KEY_ID: !Ref Key
          TABLE_NAME: !Ref UsersTable
      Policies:
        - DynamoDBReadPolicy:
            TableName: !Ref UsersTable
        - KMSDecryptPolicy:
            KeyId: !Ref Key
    Type: AWS::Serverless::Function
```
> From [template.yml](../template.yml)

The worker functions also have fine-grained policies. One has permission to create objects via the `S3CrudPolicy` and one has permission to only list and delete objects via a custom statement. Both have a `BUCKET` environment variable so the functions know what resource to use.

```yaml
Resources:
  Bucket:
    Type: AWS::S3::Bucket

  WorkerFunction:
    Properties:
      Environment:
        Variables:
          BUCKET: !Ref Bucket
      Policies:
        - S3CrudPolicy:
            BucketName: !Ref Bucket
    Type: AWS::Serverless::Function

  WorkerPeriodicFunction:
    Properties:
      Environment:
        Variables:
          BUCKET: !Ref Bucket
      Policies:
        - Statement:
            - Action:
                - s3:DeleteObject
                - s3:ListObjects
              Effect: Allow
              Resource: !Sub "arn:aws:s3:::${Bucket}/*"
    Type: AWS::Serverless::Function
```
> From [template.yml](../template.yml)

The security implications are massive.

Our users API has limited the sensitive operation of decrypting data to a single API endpoint.

Our worker functions don't know the name of the database nor do they have any DynamoDB permissions at all. We could now perform un-trusted work like running a user-supplied script with confidence it can never access our database.

Consider a common alternative where an app has a single set of IAM keys via `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables and a liberal policy like `AdministratorAccess`. A vulnerability in any one endpoint would expose keys to the entire AWS kingdom.

## SAM Policy Templates

Note how the `Policies` section has two different formats for writing policies.

In some cases we specify a `Statement`, a full IAM policy body. The [IAM policies docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_statement.html) and [CloudFormation IAM role docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-role.html) offer guidance for crafting these statements.

In other cases we specify simple statement like `DynamoDBCrudPolicy` or `S3CrudPolicy` scoped to a single resource. This is a feature of SAM called "policy templates".

A SAM template is a dialect of a CloudFormation template that makes it simpler to write the configuration for a serverless app. When we deploy this template a "transform" takes place behind the scenes, and turns our simplified policy config into a full IAM policy body.

The [SAM Policy Templates doc](https://github.com/awslabs/serverless-application-model/blob/master/docs/policy_templates.rst) lists what policy templates are available and how they are transformed.

<details>
<summary>See a transformed AWS::IAM::Role resource...</summary>
&nbsp;

```json
{
    "Resources": {
        "UserCreateFunctionRole": {
            "Type": "AWS::IAM::Role",
            "Properties": {
                "ManagedPolicyArns": [
                    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
                    "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
                ],
                "Policies": [
                    {
                        "PolicyName": "UserCreateFunctionRolePolicy0",
                        "PolicyDocument": {
                            "Statement": [
                                {
                                    "Action": [
                                        "dynamodb:GetItem",
                                        "dynamodb:DeleteItem",
                                        "dynamodb:PutItem",
                                        "dynamodb:Scan",
                                        "dynamodb:Query",
                                        "dynamodb:UpdateItem",
                                        "dynamodb:BatchWriteItem",
                                        "dynamodb:BatchGetItem"
                                    ],
                                    "Resource": {
                                        "Fn::Sub": [
                                            "arn:${AWS::Partition}:dynamodb:${AWS::Region}:${AWS::AccountId}:table/${tableName}",
                                            {
                                                "tableName": {
                                                    "Ref": "UsersTable"
                                                }
                                            }
                                        ]
                                    },
                                    "Effect": "Allow"
                                }
                            ]
                        }
                    },
                    {
                        "PolicyName": "UserCreateFunctionRolePolicy2",
                        "PolicyDocument": {
                            "Version": "2012-10-17",
                            "Statement": [
                                {
                                    "Action": [
                                        "kms:Encrypt"
                                    ],
                                    "Resource": {
                                        "Fn::GetAtt": [
                                            "Key",
                                            "Arn"
                                        ]
                                    },
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
        }
    }
}
```
</details>

## Summary

When building an app as a set of functions we now can:

- Set a separate environment for every function
- Apply separate policies to every function
- Follow the principal of least privilege

We no longer have to worry about:

- Shared-all app secrets
- Shared-all app policies

Lambda makes building apps significantly more secure.