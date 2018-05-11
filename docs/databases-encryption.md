# Databases and Encryption
### With Go, DynamoDB and KMS

Every meaningful application has to deal with state. We introduced S3 to save reports, but what do we do for CRUD -- data we will create, read, update and delete with random access? SAM offers strong opinion about the solution -- an `Amazon::Serverless::SimpleTable` resource -- which is a DynamoDB table.

The choice of a data store can make or break an application and its architecture.

Imagine what happens as our app takes off, and all of a sudden we start to get 100s of calls per second to our User API. FaaS is well-suited to this scenario -- we expect AWS to instantly scale up the 100s of Go functions for us. But this scenario -- 100s of clients simultaneously reading and writing data -- could pose a challenge for our database.

PostgreSQL, one of the goto databases for web apps, may not handle 100 simultaneous connections without adding connection pooling, or may require migrating data to a higher capacity server. Both are heavy operational tasks.

DynamoDB is better suited to this challenge.

It is highly available (HA) out of the box, which means Amazon is constantly protecting us from a single server failure taking our data offline. It has an HTTP API, which is inherently suitable for many clients requesting data simultaneously. It offers strong consistency, which is a killer feature for distributed systems at scale. It scales transparently, which means we can increase the provisioned read or write throughput and our functions can still read and write data while the database scales. DynamoDB even offers auto scaling, where AWS will automatically increase throughput when more read and write requests appear, and decrease it when they go away.

For all these reasons DynamoDB feels like a good choice for our "serverless" app.

But of course it poses some challenges.

DynamoDB not as easy to use as a developer. It lacks transactions so if we need to update multiple records atomically, our code has to handle locking, updating, then unlocking. It has a simplistic indexing model so we have to design our table keys and limited indexes carefully to avoid scanning the entire table. It's scaling model isn't perfect, so there are scenarios where DynamoDB will be inefficient and expensive at medium to large scale.

Perhaps [AWS Aurora Serverless](https://aws.amazon.com/blogs/aws/in-the-works-amazon-aurora-serverless/) will offer the best of both worlds: a SQL database that handles many connections and scales transparently. However, at time of writing it is in preview, so not available for day-to-day use.

So for many use-cases DynamoDB is indeed a "simple table" and a good default choice to add to our app.

With any data store, the strategy for storing sensitive data like API keys, credit cards, or personal information is uncontroversial. We opt to use the AWS Key Management Service (KMS) to encrypt data.

## AWS Config

SAM makes it easy to create a table and key and attach them to our functions:

```yaml
Resources:
  Key:
    Properties:
      KeyPolicy:
        Id: default
        Statement:
          - Action: kms:*
            Effect: Allow
            Principal:
              AWS: !Sub arn:aws:iam::${AWS::AccountId}:root
            Resource: '*'
            Sid: Enable IAM User Permissions
        Version: 2012-10-17
    Type: AWS::KMS::Key

  UsersTable:
    Properties:
      ProvisionedThroughput:
        ReadCapacityUnits: 1
        WriteCapacityUnits: 1
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

Here we start with the lowest value for read and write capacity units to save money. But we can consider making these parameters on the stack or adding more resources to perform autoscaling ([docs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-dynamodb-table.html#cfn-dynamodb-table-examples-application-autoscaling))...

Note how we give one function a policy to encrypt and another a policy to decrypt. This is the "the principal of least privilege". We might consider custom statements with `dynamodb:GetItem`, `dynamodb:DeleteItem`, and `dynamodb:PutItem` actions too, but we opt for the simpler template policy for now. See the [Per-Function Policies](per-function-policies.md) doc for more details.

## Go Code

Now we can encrypt, save and retrieve data with KMS and DynamoDB APIs.

```go
package gofaas

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
)

func userGet(ctx context.Context, id string, decrypt bool) (*User, error) {
	out, err := DynamoDB().GetItemWithContext(ctx, &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(id),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if out.Item == nil {
		return nil, ResponseError{"not found", 404}
	}

	u := User{
		ID:       *out.Item["id"].S,
		Token:    out.Item["token"].B,
		Username: *out.Item["username"].S,
	}

	// optionally decrypt the token ciphertext
	if decrypt {
		out, err := KMS().DecryptWithContext(ctx, &kms.DecryptInput{
			CiphertextBlob: u.Token,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		u.TokenPlain = string(out.Plaintext)
	}

	return &u, nil
}

func userPut(ctx context.Context, u *User) error {
	if u.TokenPlain != "" {
		out, err := KMS().EncryptWithContext(ctx, &kms.EncryptInput{
			Plaintext: []byte(u.TokenPlain),
			KeyId:     aws.String(os.Getenv("KEY_ID")),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		u.Token = out.CiphertextBlob
		u.TokenPlain = ""
	}

	_, err := DynamoDB().PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(u.ID),
			},
			"token": &dynamodb.AttributeValue{
				B: u.Token,
			},
			"username": &dynamodb.AttributeValue{
				S: aws.String(u.Username),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
	return errors.WithStack(err)
}
```
> From [user.go](../user.go)

Encrypting data before saving it to the database is a security best practice. If someone was to gain access to the database or a dump of data, they would not be able to access our sensitive information without gaining additional access to KMS. KMS makes this easy with its Encrypt and Decrypt APIs.

## Summary

When building an app with Go, DynamoDB and KMS we:

- Store and access data with fast, random access
- Save data in an encrypted format
- Replicate our data across multiple servers transparently
- Scale our database up or down without downtime

We don't have to:

- Operate database servers or clusters
- Design custom encryption schemes

DynamoDB and KMS make it easy to store data in a secure and reliable manner.