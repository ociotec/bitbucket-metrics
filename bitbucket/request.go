package bitbucket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type Request struct {
	BaseURL          *url.URL
	Username         string
	Password         string
	PageSize         int
	BitbucketVersion string
}

func NewRequest(baseURLString, username, password string, pageSize int) (*Request, error) {
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		log.WithFields(log.Fields{
			"base-url": baseURLString,
		}).Error("Invalid base URL to parse")
		return nil, err
	}
	request := &Request{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		PageSize: pageSize,
	}
	return request, nil
}

func (request *Request) Run(verb string, subURIs ...string) (map[string]any, error) {
	return request.RunWithArgs(verb, nil, subURIs...)
}

func (request *Request) RunWithArgs(verb string, args map[string]any, subURIs ...string) (map[string]any, error) {
	// Create the URL joining base URL + all received sub URIs
	url := *(request.BaseURL)
	url = *url.JoinPath(subURIs...)
	// Add query parameters to the URL
	query := request.BaseURL.Query()
	for name, value := range args {
		query.Set(name, fmt.Sprint(value))
	}
	url.RawQuery = query.Encode()
	// Create the HTTP request
	httpRequest, err := http.NewRequest(verb, url.String(), nil)
	if err != nil {
		log.WithFields(log.Fields{
			"verb":  verb,
			"url":   url.String(),
			"error": err,
		}).Error("Cannot create new HTTP request")
		return nil, err
	}
	// Now add headers, including authentication
	auth := request.Username + ":" + request.Password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	httpRequest.Header.Add("Authorization", "Basic "+encodedAuth)
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("charset", "UTF-8")
	// Create the HTTP client and do the request to get a response
	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"verb":  verb,
			"url":   url.String(),
			"error": err,
		}).Error("Cannot do HTTP request")
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Check OK status code
	if httpResponse.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{
			"verb":        verb,
			"url":         url.String(),
			"code-status": httpResponse.StatusCode,
		}).Error("Unexpected HTTP status code")
		return nil, err
	}

	// Read & parse the response body
	rawBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"verb":  verb,
			"url":   url.String(),
			"error": err,
		}).Error("Cannot read HTTP response body")
		return nil, err
	}
	body := string(rawBody)

	var bodyJSON map[string]any
	err = json.Unmarshal([]byte(body), &bodyJSON)
	if err != nil {
		log.WithFields(log.Fields{
			"verb":        verb,
			"url":         url.String(),
			"code-status": httpResponse.StatusCode,
			"body":        body,
		}).Error("Cannot parse JSON body")
		return nil, err
	}

	log.WithFields(log.Fields{
		"verb":        verb,
		"url":         url.String(),
		"code-status": httpResponse.StatusCode,
		"body-json":   bodyJSON,
	}).Debug("Request was successfully executed")

	return bodyJSON, nil
}
