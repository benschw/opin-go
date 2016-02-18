package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var _ = log.Print

var (
	ErrStatusConflict            error = errors.New(fmt.Sprintf("%d: Conflict", http.StatusConflict))
	ErrStatusBadRequest          error = errors.New(fmt.Sprintf("%d: Bad Request", http.StatusBadRequest))
	ErrStatusInternalServerError error = errors.New(fmt.Sprintf("%d: Internal Server Error", http.StatusInternalServerError))
	ErrStatusNotFound            error = errors.New(fmt.Sprintf("%d: Not Found", http.StatusNotFound))
)

func NewRequestH(method string, url string, headers map[string]interface{}, entity interface{}) (*http.Response, error) {
	req, err := BuildRequest(method, url, headers, entity)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func MakeRequest(method string, url string, entity interface{}) (*http.Response, error) {
	headers := map[string]interface{}{}
	req, err := BuildRequest(method, url, headers, entity)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func BuildRequest(method string, url string, headers map[string]interface{}, entity interface{}) (*http.Request, error) {
	body, err := encodeEntity(entity)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	for header, value := range headers {
		switch value.(type) {
		case int:
			req.Header.Set(header, string(value.(int)))
		case string:
			req.Header.Set(header, value.(string))
		}
	}
	return req, err
}

func encodeEntity(entity interface{}) (io.Reader, error) {
	if entity == nil {
		return nil, nil
	} else {
		b, err := json.Marshal(entity)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(b), nil
	}
}

func ProcessResponseBytes(r *http.Response, expectedStatus int) ([]byte, error) {
	if err := processResponse(r, expectedStatus); err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(r.Body)
	return respBody, err
}
func ProcessResponseEntity(r *http.Response, entity interface{}, expectedStatus int) error {
	if err := processResponse(r, expectedStatus); err != nil {
		return err
	}
	return ForceProcessResponseEntity(r, entity)
}
func ForceProcessResponseEntity(r *http.Response, entity interface{}) error {
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if entity != nil {
		if err = json.Unmarshal(respBody, entity); err != nil {
			return err
		}
	}
	return nil
}
func processResponse(r *http.Response, expectedStatus int) error {
	if r == nil {
		return errors.New("response is nil")
	}
	if r.StatusCode != expectedStatus {

		switch r.StatusCode {
		case http.StatusConflict:
			return ErrStatusConflict
		case http.StatusBadRequest:
			return ErrStatusBadRequest
		case http.StatusInternalServerError:
			return ErrStatusInternalServerError
		case http.StatusNotFound:
			return ErrStatusNotFound
		default:
			return errors.New("response status of " + r.Status)
		}

	}

	return nil
}
