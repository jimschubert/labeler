package model

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

// test helper which reads in test data by filename
func helperTestData(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
