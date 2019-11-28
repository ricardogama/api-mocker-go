package mocker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Request is a request that can be expected by the mock server.
type Request struct {
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Times    int               `json:"times,omitempty"`
	Query    map[string]string `json:"query,omitempty"`
	Response *Response         `json:"response"`
}

// Response is a response that can be expected as a response to a mock request.
type Response struct {
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	Status  int               `json:"status"`
}

// Results represents the results from the mock server.
type Results struct {
	Expected   []*Request `json:"expected"`
	Unexpected []*Request `json:"unexpected"`
}

// Mocker allows to communicate with a mock server.
type Mocker struct {
	BasePath string
}

// New returns a new Mocker.
func New(basePath string) *Mocker {
	return &Mocker{
		basePath,
	}
}

// Results returns both the expected and the unexpected results.
func (ms *Mocker) Results() (*Results, error) {
	results := &Results{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/mocks", ms.BasePath), nil)
	if err != nil {
		return nil, err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != 200 {
		return nil, errors.New("failed to get mocks")
	}

	if err := DecodeResponse(rsp, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Ensure ensures that no expected requests were not matched, and that there were no unexpected requests.
func (ms *Mocker) Ensure() error {
	results, err := ms.Results()
	if err != nil {
		return err
	}

	var errs []string

	if len(results.Expected) > 0 {
		expected, err := JSONString(results.Expected)
		if err != nil {
			return err
		}

		errs = append(errs, fmt.Sprintf("missing %d expected calls: %v", len(results.Expected), expected))
	}

	if len(results.Unexpected) > 0 {
		unexpected, err := JSONString(results.Unexpected)
		if err != nil {
			return err
		}

		errs = append(errs, fmt.Sprintf("%d unexpected calls: %v", len(results.Unexpected), unexpected))
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}

// Expect tells the actual mock server to expect the given request.
func (ms *Mocker) Expect(mock *Request) error {
	body, err := json.Marshal(mock)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/mocks", ms.BasePath), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/json")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode == 201 {
		return nil
	}

	errs, err := JSONResponse(rsp)
	if err != nil {
		return err
	}

	return fmt.Errorf("failed to create mock %s", errs)
}

// Clear tells the actual mock server to clear all expected and unexpected requests.
func (ms *Mocker) Clear() error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/mocks", ms.BasePath), nil)
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != 204 {
		return errors.New("failed to clear mocks")
	}

	return nil
}

// DecodeResponse decodes the response body into the given data.
func DecodeResponse(rsp *http.Response, data interface{}) error {
	defer rsp.Body.Close()

	return json.NewDecoder(rsp.Body).Decode(&data)
}

// JSONResponse returns the response body as a JSON string.
func JSONResponse(rsp *http.Response) (string, error) {
	defer rsp.Body.Close()

	raw, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	json.Indent(&buf, raw, "", "\t") // nolint: errcheck

	return buf.String(), nil
}

// JSONString returns the given data as a JSON string.
func JSONString(data interface{}) (string, error) {
	raw, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	json.Indent(&buf, raw, "", "\t") // nolint: errcheck

	return buf.String(), nil
}
