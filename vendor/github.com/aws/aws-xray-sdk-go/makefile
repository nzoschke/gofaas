SDK_BASE_FOLDERS=$(shell ls -d */ | grep -v vendor)
GO_VET_CMD=go tool vet --all -shadow

assets:
	rm resources/bindata.go
	go-bindata -o resources/bindata.go -pkg resources resources/

vet:
	${GO_VET_CMD} ${SDK_BASE_FOLDERS}

lint:
	golint ${SDK_BASE_FOLDERS}

test:
	go test -cover `go list ./... | grep -v vendor`

fmt:
	go fmt `go list ./... | grep -v vendor`
