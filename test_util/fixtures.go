package test_util

import (
	"testing"
	"runtime"
	"path/filepath"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"github.com/pkg/errors"
)

func LoadFixture(t *testing.T, out interface{}) error {
	_, caller, _, _ := runtime.Caller(1)
	dir := filepath.Dir(caller)
	filename := fmt.Sprintf("%s.json", t.Name())
	fullpath := filepath.Join(dir, "testdata", filename)
	data, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, out)
	if err != nil {
		return errors.Cause(err)
	}
	return nil
}
