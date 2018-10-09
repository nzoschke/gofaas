export AWS_DEFAULT_REGION ?= us-east-1
APP ?= gofaas

app: dev

clean:
	rm -f $(wildcard handlers/*/main)
	rm -rf $(wildcard web/handlers/*/node_modules)

deploy: BUCKET = pkgs-$(shell aws sts get-caller-identity --output text --query 'Account')-$(AWS_DEFAULT_REGION)
deploy: PARAMS ?= =
deploy: handlers
	@aws s3api head-bucket --bucket $(BUCKET) || aws s3 mb s3://$(BUCKET) --region $(AWS_DEFAULT_REGION)
	sam package --output-template-file out.yml --s3-bucket $(BUCKET) --template-file template.yml
	sam deploy --capabilities CAPABILITY_NAMED_IAM --parameter-overrides $(PARAMS) --template-file out.yml --stack-name $(APP)
	make deploy-static

deploy-static: API_URL=$(shell aws cloudformation describe-stacks --output text --query 'Stacks[].Outputs[?OutputKey==`ApiUrl`].{Value:OutputValue}' --stack-name $(APP))
deploy-static: BUCKET=$(shell aws cloudformation describe-stack-resources --output text --query 'StackResources[?LogicalResourceId==`WebBucket`].{Id:PhysicalResourceId}' --stack-name $(APP))
deploy-static: DIST=$(shell aws cloudformation describe-stack-resources --output text --query 'StackResources[?LogicalResourceId==`WebDistribution`].{Id:PhysicalResourceId}' --stack-name $(APP))
deploy-static: web/static/index.html
	echo "const API_URL=\"$(API_URL)\";" > web/static/js/env.js
	aws s3 sync web/static s3://$(BUCKET)/
	[ -n "$(DIST)" ] && aws cloudfront create-invalidation --distribution-id $(DIST) --paths '/*' || true
	aws cloudformation describe-stacks --output text --query 'Stacks[*].Outputs' --stack-name $(APP)

dev-debug:
	make clean
	GCFLAGS="-N -l" make -j handlers
	GOARCH=amd64 GOOS=linux go build -o /tmp/delve/dlv github.com/derekparker/delve/cmd/dlv
	sam local start-api -d 5986 --debugger-path /tmp/delve -n env.json -s web/static

dev:
	make -j dev-watch dev-sam
dev-sam:
	sam local start-api -n env.json -s web/static
dev-watch:
	watchexec -f '*.go' 'make -j handlers'

HANDLERS=$(addsuffix main,$(wildcard handlers/*/))
$(HANDLERS): handlers/%/main: *.go handlers/%/main.go
	cd ./$(dir $@) && GOOS=linux go build -gcflags="${GCFLAGS}" -o main .

HANDLERS_JS=$(addsuffix node_modules,$(wildcard web/handlers/*/))
$(HANDLERS_JS): web/handlers/%/node_modules: web/handlers/%/package.json
	cd ./$(dir $@) && npm install && node-prune >/dev/null && touch node_modules

handlers: handlers-go handlers-js
handlers-go: $(HANDLERS)
handlers-js: $(HANDLERS_JS)

test:
	go test -v ./...
