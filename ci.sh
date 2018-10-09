#!/bin/bash
set -e

trap 'echo 🈲 ERROR' ERR

RAND=$RANDOM
export APP=gofaas-$RAND
export AWS_PROFILE=gofaas

make -j deploy

API_URL=$(aws cloudformation describe-stacks --output text --query 'Stacks[].Outputs[?OutputKey==`ApiUrl`].{Value:OutputValue}' --stack-name $APP)
BUCKET=$(aws cloudformation describe-stack-resources --output text --query 'StackResources[?LogicalResourceId==`Bucket`].{Id:PhysicalResourceId}' --stack-name $APP)
WEB_URL=$(aws cloudformation describe-stacks --output text --query 'Stacks[].Outputs[?OutputKey==`WebUrl`].{Value:OutputValue}' --stack-name $APP)
WEB_BUCKET=$(aws cloudformation describe-stack-resources --output text --query 'StackResources[?LogicalResourceId==`WebBucket`].{Id:PhysicalResourceId}' --stack-name $APP)

# test static site
curl -s $WEB_URL | grep "My first gofaas"

# test user funcs
ID=$(curl -s -X POST $API_URL/users -d '{"username":"test"}' | jq -r .id)
curl -s $API_URL/users/$ID | grep test
curl -s $API_URL/users/$ID?token=true | grep token
curl -s -d '{"username": "test2"}' -X PUT $API_URL/users/$ID | grep test2
curl -s -X DELETE $API_URL/users/$ID | grep test2
curl -s $API_URL/users/$ID | grep "not found"

# test worker API and funcs
curl -s -X POST $API_URL/work | grep 202
sleep 2
[ $(aws s3 ls s3://$BUCKET | wc -l) -eq "1" ] # 1 file in bucket
aws lambda invoke --function-name $APP-WorkerPeriodicFunction --log-type Tail /dev/null | grep 200
[ $(aws s3 ls s3://$BUCKET | wc -l) -eq "0" ] # 0 files in bucket

# TODO: Test private mode. Currently hard with cert approval and DNS
# export AUTH_HASH_KEY=43Z647ntcQ8L5LfNi2HlW3XXJYz5x9Y/EYv6C7gdajo=
# export ACCESS_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0QGdvZmFhcy5uZXQiLCJleHAiOjIwMDAwMDAwMDB9.8I4HeBoWs1rcXDclctz2qJaTrrRHm0aKZOCJtMfwaQE
# make deploy PARAMS="ApiDomainName=api-$RAND.gofaas.net AuthDomainName=gofaas.net AuthHashKey=$AUTH_HASH_KEY OAuthClientId=foo OAuthClientSecret=bar WebDomainName=www-$RAND.gofaas.net"

aws s3 rm --recursive s3://$WEB_BUCKET/
aws cloudformation delete-stack --stack-name $APP
aws cloudformation wait stack-delete-complete --stack-name $APP
echo ✅ SUCCESS!