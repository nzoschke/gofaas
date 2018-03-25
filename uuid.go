package gofaas

import (
	"github.com/satori/go.uuid"
)

// UUIDGen is a UUID generator that can be mocked for testing
var UUIDGen = func() uuid.UUID {
	return uuid.NewV4()
}
