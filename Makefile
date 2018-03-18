export AWS_DEFAULT_REGION ?= us-east-1
APP = gofaas
PKG = github.com/nzoschke/$(APP)

app: dev

clean:
	rm -f $(wildcard handlers/*/main)
	rm -f $(wildcard handlers/*/main.zip)
	rm -f web-auth/index.zip

deploy: BUCKET = pkgs-$(shell aws sts get-caller-identity --output text --query 'Account')-$(AWS_DEFAULT_REGION)
deploy: PARAMS ?= =
deploy: handlers web-auth/index.zip
	@aws s3api head-bucket --bucket $(BUCKET) || aws s3 mb s3://$(BUCKET) --region $(AWS_DEFAULT_REGION)
	aws cloudformation package --output-template-file out.yml --s3-bucket $(BUCKET) --template-file template.yml
	aws cloudformation deploy --capabilities CAPABILITY_NAMED_IAM --parameter-overrides $(PARAMS) --template-file out.yml --stack-name $(APP)
	aws cloudformation describe-stacks --output text --query 'Stacks[*].Outputs' --stack-name $(APP)

deploy-static: BUCKET=$(shell aws cloudformation describe-stack-resources --output text --query 'StackResources[?LogicalResourceId==`WebBucket`].{Id:PhysicalResourceId}' --stack-name $(APP))
deploy-static: DIST=$(shell aws cloudformation describe-stack-resources --output text --query 'StackResources[?LogicalResourceId==`WebDistribution`].{Id:PhysicalResourceId}' --stack-name $(APP))
deploy-static: public/index.html
	aws s3 sync public s3://$(BUCKET)/
	[ -n "$(DIST)" ] && aws cloudfront create-invalidation --distribution-id $(DIST) --paths '/*' || true
	aws cloudformation describe-stacks --output text --query 'Stacks[*].Outputs' --stack-name $(APP)

dev:
	make -j dev-watch dev-sam
dev-sam:
	aws-sam-local local start-api -n env.json
dev-watch:
	watchexec -f '*.go' 'make -j handlers'

HANDLERS=$(addsuffix main.zip,$(wildcard handlers/*/))
handlers: $(HANDLERS)
$(HANDLERS): handlers/%/main.zip: *.go handlers/%/main.go
	cd ./$(dir $@) && GOOS=linux go build -o main . && zip -1r -xmain.go main.zip *

web-auth/index.zip: web-auth/*.js
	go get github.com/tj/node-prune/cmd/node-prune
	cd ./web-auth && npm install && node-prune && zip -9r index.zip *

test:
	go test -v ./...
