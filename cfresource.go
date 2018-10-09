package gofaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

// CFEvent is a CloudFormation Custom Resource Event that Lambda is invoked with
type CFEvent struct {
	LogicalResourceID     string          `json:"LogicalResourceId"`
	OldResourceProperties json.RawMessage `json:"OldResourceProperties"`
	PhysicalResourceID    string          `json:"PhysicalResourceId"`
	RequestID             string          `json:"RequestId"`
	RequestType           string          `json:"RequestType"`
	ResourceProperties    json.RawMessage `json:"ResourceProperties"`
	ResourceType          string          `json:"ResourceType"`
	ResponseURL           string          `json:"ResponseURL"`
	StackID               string          `json:"StackId"`
}

// CFResponse is a CloudFormation Custom Resource Response that is PUT to S3
type CFResponse struct {
	Data struct {
		Value string `json:"Value"`
	}
	LogicalResourceID  string `json:"LogicalResourceId"`
	PhysicalResourceID string `json:"PhysicalResourceId"`
	Reason             string `json:"Reason"`
	RequestID          string `json:"RequestId"`
	StackID            string `json:"StackId"`
	Status             string `json:"Status"`
}

// CFRespond creates, updates or deletes a CloudFormation custom resource
// then PUTs the status to the given ResponseURL
func CFRespond(ctx context.Context, e CFEvent) error {
	fmt.Printf("EVENT: %+v\n", e)

	r := CFResponse{
		LogicalResourceID: e.LogicalResourceID,
		RequestID:         e.RequestID,
		StackID:           e.StackID,
	}

	id, err := CFResource(ctx, e)
	if err != nil {
		r.Reason = err.Error()
		r.Status = "FAILED"
	} else {
		r.PhysicalResourceID = id
		r.Status = "SUCCESS"
	}

	body, err := json.Marshal(r)
	if err != nil {
		return err
	}
	log.Printf("CF RESPONSE:\n%s", body)

	req, err := http.NewRequest(http.MethodPut, e.ResponseURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Del("Content-Type")

	cln := &http.Client{}
	res, err := cln.Do(req)
	if err != nil {
		return err
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	log.Printf("CF STATUS: %d %s\n %s", res.StatusCode, res.Status, body)
	return nil
}

// CFResource dispatches a CloudFormation Custom Resource event to a create, update or delete function
func CFResource(ctx context.Context, e CFEvent) (string, error) {
	var props map[string]string
	if err := json.Unmarshal(e.ResourceProperties, &props); err != nil {
		return "", err
	}

	if props["RestApiId"] == "" {
		return "", fmt.Errorf("RestApiId is required")
	}

	if props["Stage"] == "" {
		return "", fmt.Errorf("Stage is required")
	}

	if props["TracingEnabled"] == "" {
		return "", fmt.Errorf("TracingEnabled is required")
	}

	switch e.RequestType {
	case "Create":
		return CFResourceCreate(ctx, e, props)
	case "Update":
		return CFResourceUpdate(ctx, e, props)
	case "Delete":
		return CFResourceDelete(ctx, e, props)
	}

	return "", fmt.Errorf("Unknown RequestType %s", e.RequestType)
}

// CFResourceCreate creates a custom resource and returns a new PhysicalResourceID
// This delegates to ResourceUpdate since there are no resources to create
func CFResourceCreate(ctx context.Context, e CFEvent, props map[string]string) (string, error) {
	e.PhysicalResourceID = resourceID(e)
	return CFResourceUpdate(ctx, e, props)
}

// CFResourceUpdate updates a custom resource and returns its existing PhysicalResourceID
// This toggles settings on the Stage specified by Stage and RestApiId parameters
func CFResourceUpdate(ctx context.Context, e CFEvent, props map[string]string) (string, error) {
	_, err := APIGateway.UpdateStageWithContext(ctx, &apigateway.UpdateStageInput{
		PatchOperations: []*apigateway.PatchOperation{
			&apigateway.PatchOperation{
				Op:    aws.String("replace"),
				Path:  aws.String("/tracingEnabled"),
				Value: aws.String(props["TracingEnabled"]),
			},
		},
		RestApiId: aws.String(props["RestApiId"]),
		StageName: aws.String(props["Stage"]),
	})
	return e.PhysicalResourceID, err
}

// CFResourceDelete deletes a custom resource and returns its old PhysicalResourceID
// This delegates to ResourceUpdate since there are no resources to delete
func CFResourceDelete(ctx context.Context, e CFEvent, props map[string]string) (string, error) {
	props["TracingEnabled"] = "false"
	return CFResourceUpdate(ctx, e, props)
}

// resourceID generates id of the form "StackName-LogicalResourceID-RandomID".
func resourceID(e CFEvent) string {
	rns := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	snm := strings.Split(e.StackID, "/")[1]
	lid := e.LogicalResourceID
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd := make([]byte, 12)
	for i := range rnd {
		rnd[i] = rns[gen.Intn(len(rns))]
	}
	return fmt.Sprintf("%s-%s-%s", snm, lid, rnd)
}
