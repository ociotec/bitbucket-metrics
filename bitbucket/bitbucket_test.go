package bitbucket

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitWithValidVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Invalid verb '%v', expected GET", r.Method)
		}
		expectedURL := fmt.Sprintf("/%s/application-properties", API_PATH)
		if r.URL.String() != expectedURL {
			t.Errorf("Invalid URL '%v', expected '%v'", r.URL.String(), expectedURL)
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
		w.Write([]byte("{\"version\": \"1.2.3\"}"))
	}))
	defer ts.Close()

	req := Init(ts.URL, "username", "password", 123)
	if req == nil {
		t.Error("Init failed")
	} else if req.BitbucketVersion != "1.2.3" {
		t.Errorf("Bitbucket version '%v' is not the expected '1.2.3'", req.BitbucketVersion)
	}
}

func TestInitWithInvalidURL(t *testing.T) {
	// We defer this anonymous function to recover from a panic call
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on Init with an invalid URL")
		}
	}()
	Init("\tinvalid", "", "", 1)
}

func TestInitWithInvalidResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("no-valid-JSON"))
	}))
	defer ts.Close()

	// We defer this anonymous function to recover from a panic call
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on Init with an invalid JSON response without version field")
		}
	}()

	Init(ts.URL, "username", "password", 123)
}

func TestInitWithValidResponseButNoVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"no-version\": \"1.2.3\"}"))
	}))
	defer ts.Close()

	// We defer this anonymous function to recover from a panic call
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on Init with an invalid JSON response without version field")
		}
	}()

	Init(ts.URL, "username", "password", 123)
}
