package client_test

import (
	"net/http"
	"reflect"
	"testing"
)

func assertHeader(t *testing.T, response testResponse, key string, val []string) {
	contentTypeHeader := response.Header[http.CanonicalHeaderKey(key)]
	if !reflect.DeepEqual(contentTypeHeader, val) {
		t.Fatalf("expected response header to be '%#v', got '%#v'", val, contentTypeHeader)
	}
}

func assertMethod(t *testing.T, response testResponse, method string) {
	if response.Method != method {
		t.Fatalf("expected response method to be '%s', got '%s'", method, response.Method)
	}
}

func assertPath(t *testing.T, response testResponse, path string) {
	if response.Path != path {
		t.Fatalf("expected response path to be '%s', got '%s'", path, response.Path)
	}
}
