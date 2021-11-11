package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KompiTech/rmap"
)

const (
	couchDBEndpointEnv = "COUCHDB_ENDPOINT"
	couchDBNameEnv     = "COUCHDB_DBNAME"
	couchDBUserEnv     = "COUCHDB_USER"
	couchDBPassEnv     = "COUCHDB_PASS"
)

type CouchDBClient struct {
	endpoint   string
	DBName     string
	httpClient http.Client
	authCookie *http.Cookie
}

func NewCouchDBClientFromEnv() (*CouchDBClient, error) {
	endpoint, err := mandatoryEnv(couchDBEndpointEnv)
	if err != nil {
		return nil, err
	}

	dbName, err := mandatoryEnv(couchDBNameEnv)
	if err != nil {
		return nil, err
	}

	user, _ := os.LookupEnv(couchDBUserEnv)
	pass, _ := os.LookupEnv(couchDBPassEnv)

	return NewCouchDBClient(endpoint, dbName, user, pass)
}

func NewCouchDBClient(endpoint, dbName, user, pass string) (*CouchDBClient, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cdb := &CouchDBClient{
		endpoint:   endpoint,
		DBName:     dbName,
		httpClient: http.Client{Transport: tr, Timeout: 5 * time.Second},
	}

	if user != "" {
		if err := cdb.auth(user, pass); err != nil {
			return nil, err
		}
	}

	return cdb, nil
}

func (c *CouchDBClient) auth(user, pass string) error {
	data := rmap.NewFromMap(map[string]interface{}{
		"name":     user,
		"password": pass,
	})

	url := c.endpoint + "/_session"

	req, err := http.NewRequest("POST", url, bytes.NewReader(data.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Add("content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("POST %s %d, response body: %s", url, resp.StatusCode, body)
	}

	c.authCookie = resp.Cookies()[0]
	return nil
}

func (c *CouchDBClient) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.endpoint+path, body)
	if err != nil {
		return nil, err
	}

	if c.authCookie != nil {
		req.AddCookie(c.authCookie)
	}

	if body != nil {
		req.Header.Add("content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *CouchDBClient) getPath(key string) string {
	return "/" + c.DBName + "/" + key
}

func (c *CouchDBClient) put(key, rev, valueJson string) (string, error) {
	if rev != "" {
		data, err := rmap.NewFromString(valueJson)
		if err != nil {
			return "", err
		}

		data.Mapa["_rev"] = rev
		valueJson = data.String()
	}

	path := c.getPath(key)
	resp, err := c.makeRequest("PUT", path, strings.NewReader(valueJson))
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("PUT %s %d, response body: %s", path, resp.StatusCode, body)
	}

	return getRev(resp.Header), nil
}

// Insert will fail if document with key already exists
func (c *CouchDBClient) Insert(key, valueJson string) error {
	_, err := c.put(key, "", valueJson)
	return err
}

// Update will fail if document does not exist
func (c *CouchDBClient) Update(key, valueJson string) error {
	path := c.getPath(key)
	headResp, err := c.makeRequest("HEAD", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = headResp.Body.Close() }()

	if headResp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(headResp.Body)
		return fmt.Errorf("HEAD %s %d, response body: %s", path, headResp.StatusCode, errBody)
	}

	_, err = c.put(key, getRev(headResp.Header), valueJson)
	return err
}

// InsertIfNotExist will create a new document if not exist
func (c *CouchDBClient) InsertIfNotExist(key, valueJson string) error {
	path := c.getPath(key)
	headResp, err := c.makeRequest("HEAD", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = headResp.Body.Close() }()

	rev := ""
	if headResp.StatusCode == http.StatusOK {
		return nil
	} else if headResp.StatusCode == http.StatusNotFound {
		// creating a new document, leave empty rev string
	} else {
		body, _ := ioutil.ReadAll(headResp.Body)
		return fmt.Errorf("HEAD %s %d, response body: %s", path, headResp.StatusCode, body)
	}

	_, err = c.put(key, rev, valueJson)
	return err
}

// Upsert will create a new document, or update existing
func (c *CouchDBClient) Upsert(key, valueJson string) error {
	path := c.getPath(key)
	headResp, err := c.makeRequest("HEAD", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = headResp.Body.Close() }()

	rev := ""
	if headResp.StatusCode == http.StatusOK {
		// updating existing document, use its rev
		rev = getRev(headResp.Header)
	} else if headResp.StatusCode == http.StatusNotFound {
		// creating a new document, leave empty rev string
	} else {
		body, _ := ioutil.ReadAll(headResp.Body)
		return fmt.Errorf("HEAD %s %d, response body: %s", path, headResp.StatusCode, body)
	}

	_, err = c.put(key, rev, valueJson)
	return err
}

func getRev(hdrs http.Header) string {
	return strings.Replace(hdrs["Etag"][0], "\"", "", -1)
}

func mandatoryEnv(key string) (string, error) {
	val, isDefined := os.LookupEnv(key)
	if !isDefined {
		return "", fmt.Errorf("mandatory ENV variable: %s is not defined", key)
	}

	return val, nil
}
