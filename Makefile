export AWS_DEFAULT_REGION ?= us-west-2
APP = gofaas
PKG = github.com/nzoschke/$(APP)

app: dev

clean:
	rm -f ./handlers/dashboard/{handler,handler.zip}

deploy: BUCKET = pkgs-$(shell aws sts get-caller-identity --output text --query 'Account')-$(AWS_DEFAULT_REGION)
deploy:
	@aws s3api head-bucket --bucket $(BUCKET) || aws s3 mb s3://$(BUCKET) --region $(AWS_DEFAULT_REGION)
	aws cloudformation package --output-template-file out.yml --s3-bucket $(BUCKET) --template-file template.yml
	aws cloudformation deploy --capabilities CAPABILITY_NAMED_IAM --template-file out.yml --stack-name $(APP)
	aws cloudformation describe-stacks --output text --query 'Stacks[*].Outputs' --stack-name $(APP)

dev: handlers/dashboard/handler.zip handlers/worker/handler.zip
	aws-sam-local local start-api -n env.json

handlers/dashboard/handler.zip: *.go handlers/dashboard/*.go
	cd ./handlers/dashboard && GOOS=linux go build -o handler . && zip handler.zip handler

handlers/worker/handler.zip: *.go handlers/worker/*.go
	cd ./handlers/worker && GOOS=linux go build -o handler . && zip handler.zip handler

test:
	go test -v ./...