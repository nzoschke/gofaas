#!/bin/bash
set -ex

# assert handlers build
make -j handlers

# start localstack and create a bucket and table
docker run -p 4567-4583:4567-4583 localstack/localstack

export                                              \
    AWS_ACCESS_KEY_ID=foo                           \
    AWS_DEFAULT_REGION=us-east-1                    \
    AWS_SECRET_ACCESS_KEY=bar                       \
    BUCKET=gofaas-test-bucket-1gnruqgwf8024         \
    DYNAMODB_ENDPOINT=http://localhost:4569         \
    S3_ENDPOINT=http://localhost:4572               \
    TABLE_NAME=gofaas-test-UsersTable-K9H8HL8QHRI1

aws --endpoint-url=$S3_ENDPOINT \
    s3 mb s3://$BUCKET

aws --endpoint-url=$DYNAMODB_ENDPOINT                                       \
    dynamodb create-table                                                   \
        --attribute-definitions AttributeName=id,AttributeType=S            \
        --key-schema AttributeName=id,KeyType=HASH                          \
        --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1   \
        --table-name $TABLE_NAME

# start with 
aws-sam-local local start-api