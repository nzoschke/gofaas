PKG=github.com/nzoschke/gofaas

dev: handlers/dashboard/handler.zip
	aws-sam-local local start-api -n env.json

handlers/dashboard/handler.zip: *.go handlers/dashboard/*.go
	cd ./handlers/dashboard && GOOS=linux go build -o handler . && zip handler.zip handler

clean:
	rm -f ./handlers/dashboard/{handler,handler.zip}

test:
	go test -v ./...