package bitbucket

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequestValidArguments(t *testing.T) {
	r, err := NewRequest("base", "username", "password", 123)
	if err != nil {
		t.Errorf("Not expected error creating a new request: %v", err)
	}
	if r.BaseURL.String() != "base" {
		t.Errorf("Unexpected base URL in the request, got '%v' instead of 'base'", r.BaseURL)
	}
	if r.Username != "username" {
		t.Errorf("Unexpected username in the request, got '%v' instead of 'username'", r.Username)
	}
	if r.Password != "password" {
		t.Errorf("Unexpected password in the request, got '%v' instead of 'password'", r.Password)
	}
	if r.PageSize != 123 {
		t.Errorf("Unexpected page size in the request, got '%v' instead of 123", r.PageSize)
	}
}

func TestNewRequestInvalidArguments(t *testing.T) {
	_, err := NewRequest("\t", "username", "password", 123)
	if err == nil {
		t.Errorf("Expected error creating a new request")
	}
}

func TestRunWithArgsWithValidVerbPathHeadersAndJSONBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Invalid verb '%v', expected GET", r.Method)
		}
		if r.URL.String() != "/1/2/3?arg1=value1&arg2=value2" {
			t.Errorf("Invalid URL '%v', expected '/1/2/3?arg1=value1&arg2=value2'", r.URL.String())
		}
		expectedAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte("username:password"))
		if r.Header.Get("Authorization") != expectedAuthorization {
			t.Errorf("Invalid authorization header '%v', expected '%v'", r.Header.Get("Authorization"), expectedAuthorization)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Invalid content type header '%v', expected 'application/json'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("charset") != "UTF-8" {
			t.Errorf("Invalid charset header '%v', expected 'UTF-8'", r.Header.Get("charset"))
		}
		w.Write([]byte("{\"valid\": \"true\"}"))
	}))
	defer ts.Close()

	req, err := NewRequest(ts.URL, "username", "password", 123)
	if err != nil {
		t.Errorf("NewRequest failed with error: %v", err)
	}
	values, err := req.RunWithArgs("GET", map[string]any{"arg1": "value1", "arg2": "value2"}, "1", "2", "3")
	if err != nil {
		t.Errorf("Run failed with error: %v", err)
	}
	valid, okValid := values["valid"].(string)
	if !okValid {
		t.Errorf("Returned JSON is not valid: %v", values)
	}
	if valid != "true" {
		t.Errorf("Returned JSON valid value is not 'true': '%v'", valid)
	}
}

func TestRunWithValidPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/1/2/3" {
			t.Errorf("Invalid URL '%v', expected '/1/2/3'", r.URL.String())
		}
		w.Write([]byte("{\"valid\": \"true\"}"))
	}))
	defer ts.Close()

	req, err := NewRequest(ts.URL, "username", "password", 123)
	if err != nil {
		t.Errorf("NewRequest failed with error: %v", err)
	}
	_, err = req.Run("GET", "1", "2", "3")
	if err != nil {
		t.Errorf("Run failed with error: %v", err)
	}
}
