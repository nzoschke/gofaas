export AWS_DEFAULT_REGION ?= us-east-1
APP = gofaas
PKG = github.com/nzoschke/$(APP)

app: dev

clean:
	rm -f $(wildcard handlers/*/handler*)

deploy: BUCKET = pkgs-$(shell aws sts get-caller-identity --output text --query 'Account')-$(AWS_DEFAULT_REGION)
deploy: PARAMS ?= =
deploy: handlers
	@aws s3api head-bucket --bucket $(BUCKET) || aws s3 mb s3://$(BUCKET) --region $(AWS_DEFAULT_REGION)
	aws cloudformation package --output-template-file out.yml --s3-bucket $(BUCKET) --template-file template.yml
	aws cloudformation deploy --capabilities CAPABILITY_NAMED_IAM --parameter-overrides $(PARAMS) --template-file out.yml --stack-name $(APP)
	aws cloudformation describe-stacks --output text --query 'Stacks[*].Outputs' --stack-name $(APP)

dev:
	make -j dev-watch dev-sam
dev-sam:
	aws-sam-local local start-api -n env.json
dev-watch:
	watchexec -f "*.go" -n 'make -j handlers'

HANDLERS=$(addsuffix handler.zip,$(wildcard handlers/*/))
handlers: $(HANDLERS)
$(HANDLERS): handlers/%/handler.zip: *.go handlers/%/main.go
	cd ./$(dir $@) && GOOS=linux go build -o handler . && zip -1 handler.zip handler

test:
	go test -v ./...