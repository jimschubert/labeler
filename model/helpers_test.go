package model

import (
	"os"
	"path/filepath"
	"testing"
)

// test helper which reads in test data by filename
func helperTestData(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name) // relative path
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
