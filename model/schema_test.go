package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestSchemaValidation(t *testing.T) {

	c := jsonschema.NewCompiler()
	mainUrl := "https://raw.githubusercontent.com/jimschubert/labeler/HEAD/model/schema/labeler.schema.json"

	// Load the main schema
	labelerSchema, err := os.Open("schema/labeler.schema.json")
	assert.NoError(t, err, "failed to read schema: schema/labeler.schema.json")
	assert.NoError(t, c.AddResource(mainUrl, labelerSchema))

	// Load the simple schema
	simpleSchema, err := os.Open("schema/labeler.simple.schema.json")
	assert.NoError(t, err, "failed to read schema: schema/labeler.simple.schema.json")
	assert.NoError(t, c.AddResource(
		"https://raw.githubusercontent.com/jimschubert/labeler/HEAD/model/schema/labeler.simple.schema.json",
		simpleSchema,
	))

	// Load the full schema
	fullSchema, err := os.Open("schema/labeler.full.schema.json")
	assert.NoError(t, err, "failed to read schema: schema/labeler.full.schema.json")
	assert.NoError(t, c.AddResource(
		"https://raw.githubusercontent.com/jimschubert/labeler/HEAD/model/schema/labeler.full.schema.json",
		fullSchema,
	))

	schema, err := c.Compile(mainUrl)
	if err != nil {
		t.Fatalf("failed to compile schema: %v", err)
	}

	// Find all test data files
	entries, err := os.ReadDir(filepath.Join("testdata", "schemas"))
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		t.Run(fmt.Sprintf("Validate %s against JSON Schema", entry.Name()), func(t *testing.T) {
			wantErr := !strings.Contains(entry.Name(), "_valid")
			testPath := filepath.Join("testdata", "schemas", entry.Name())
			data, err := os.ReadFile(testPath)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			var c Config
			if strings.HasPrefix(entry.Name(), "full_config") {
				c = &FullConfig{}
			} else {
				c = &SimpleConfig{}
			}

			// don't use a config's FromBytes here, because it will result in nil with error for invalid schemas
			assert.NoError(t, yaml.Unmarshal(data, c), "failed to unmarshal YAML full config: %v", err)

			jsonData, err := json.Marshal(&c)

			var obj interface{}
			if err := json.Unmarshal(jsonData, &obj); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			err = schema.Validate(obj)
			if (err != nil) != wantErr {
				t.Errorf("schema validation expectation (wantErr=%t, gotErr=true) failed: %v", wantErr, err)
			}

			var validationError *jsonschema.ValidationError
			if errors.As(err, &validationError) {
				t.Logf("Validation messages:\n%+v", validationError)
			}
		})
	}
}
