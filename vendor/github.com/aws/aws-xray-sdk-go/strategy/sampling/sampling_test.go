// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalizedStrategy(t *testing.T) {
	ss, err := NewLocalizedStrategy()
	assert.NotNil(t, ss)
	assert.Nil(t, err)
}

func TestNewLocalizedStrategyFromFilePath(t *testing.T) {
	ruleString :=
		`{
	  "version": 1,
	  "default": {
	    "fixed_target": 1,
	    "rate": 0.05
	  },
	  "rules": [
	  ]
	}`
	goPath := os.Getenv("PWD")
	testFile := goPath + "/test_rule.json"
	f, err := os.Create(testFile)
	if err != nil {
		panic(err)
	}
	f.WriteString(ruleString)
	f.Close()
	ss, err := NewLocalizedStrategyFromFilePath(testFile)
	assert.NotNil(t, ss)
	assert.Nil(t, err)
	os.Remove(testFile)
}

func TestNewLocalizedStrategyFromFilePathWithInvalidJSON(t *testing.T) {
	ruleString :=
		`{
	  "version": 1,
	  "default": {
	    "fixed_target": 1,
	    "rate":
	  },
	  "rules": [
	  ]
	}`
	goPath := os.Getenv("PWD")
	testFile := goPath + "/test_rule.json"
	f, err := os.Create(testFile)
	if err != nil {
		panic(err)
	}
	f.WriteString(ruleString)
	f.Close()
	ss, err := NewLocalizedStrategyFromFilePath(testFile)
	assert.Nil(t, ss)
	assert.NotNil(t, err)
	os.Remove(testFile)
}

func TestNewLocalizedStrategyFromJSONBytes(t *testing.T) {
	ruleBytes := []byte(`{
	  "version": 1,
	  "default": {
	    "fixed_target": 1,
	    "rate": 0.05
	  },
	  "rules": [
	  ]
	}`)
	ss, err := NewLocalizedStrategyFromJSONBytes(ruleBytes)
	assert.NotNil(t, ss)
	assert.Nil(t, err)
}

func TestNewLocalizedStrategyFromInvalidJSONBytes(t *testing.T) {
	ruleBytes := []byte(`{
	  "version": 1,
	  "default": {
	    "fixed_target": 1,
	    "rate":
	  },
	  "rules": [
	  ]
	}`)
	ss, err := NewLocalizedStrategyFromJSONBytes(ruleBytes)
	assert.Nil(t, ss)
	assert.NotNil(t, err)
}
