package openbank

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

type data struct {
	Field  string `json:"field,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type responseBody struct {
	Data []data `json:"data,omitempty"`
}

func testClientDefaultURLs(t *testing.T, c *Client) {
	if c.ApiBaseURL == nil || c.ApiBaseURL.String() != prodAPIBaseURL {
		t.Errorf("NewClient ApiBaseURL = %v, expected %v", c.ApiBaseURL, prodAPIBaseURL)
	}

	if c.AccountURL == nil || c.AccountURL.String() != prodAccountURL {
		t.Errorf("NewClient AccountURL = %v, expected %v", c.AccountURL, prodAccountURL)
	}

	if c.SiteURL == nil || c.SiteURL.String() != prodSiteURL {
		t.Errorf("NewClient SiteURL = %v, expected %v", c.AccountURL, prodAccountURL)
	}
}

func testClientDefaults(t *testing.T, c *Client) {
	testClientDefaultURLs(t, c)
}

func TestNewClient(t *testing.T) {
	c, _ := NewClient()
	testClientDefaults(t, c)
}

func TestClientWithSandboxURLs(t *testing.T) {
	c, _ := NewClient(UseSandbox())
	if c.ApiBaseURL == nil || c.ApiBaseURL.String() != sandboxAPIBaseURL {
		t.Errorf("NewClient ApiBaseURL = %v, expected %v", c.ApiBaseURL, sandboxAPIBaseURL)
	}

	if c.AccountURL == nil || c.AccountURL.String() != sandboxAccountURL {
		t.Errorf("NewClient AccountURL = %v, expected %v", c.AccountURL, sandboxAccountURL)
	}

	if c.SiteURL == nil || c.SiteURL.String() != sandboxSiteURL {
		t.Errorf("NewClient SiteURL = %v, expected %v", c.AccountURL, sandboxAccountURL)
	}
}

func TestNewAPIRequest(t *testing.T) {
	c, _ := NewClient()

	inURL, outURL := "/test", prodAPIBaseURL+"/test"
	inBody, outBody := struct {
		AccountID string     `json:"account_id"`
		Items     []struct{} `json:"items"`
		Customer  struct {
			Name string `json:"name"`
		} `json:"customer"`
		Closed   bool       `json:"closed"`
		Payments []struct{} `json:"payments"`
	}{AccountID: "abc123"},
		`{"account_id":"abc123","items":null,"customer":{"name":""},"closed":false,"payments":null}`+"\n"
	req, _ := c.NewAPIRequest(http.MethodPost, inURL, inBody)

	if req.URL.String() != outURL {
		t.Errorf("NewAPIRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != outBody {
		t.Errorf("NewAPIRequest(%v)Body = %v, expected %v", inBody, string(body), outBody)
	}

	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewAPIRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

func TestCheckResponse(t *testing.T) {
	testCases := []struct {
		Name          string
		ResponseBody  responseBody
		StatusCode    int
		ExpectedError bool
		ErrorResponse *responseBody
		RequestMethod string
		RequestUrl    string
	}{
		{
			Name: "Should return error for unsuccessful status code",
			ResponseBody: responseBody{
				Data: []data{
					{
						Field:  "Test",
						Detail: "Error",
					},
				},
			},
			StatusCode:    400,
			ExpectedError: true,
			ErrorResponse: new(responseBody),
			RequestUrl:    "http://127.0.0.1:3001/test",
			RequestMethod: "GET",
		},
		{
			Name: "Should not return error for successful status code",
			ResponseBody: responseBody{
				Data: []data{
					{
						Field:  "Test",
						Detail: "Success",
					},
				},
			},
			StatusCode:    200,
			ExpectedError: false,
			ErrorResponse: new(responseBody),
			RequestUrl:    "http://127.0.0.1:3001/test",
			RequestMethod: "GET",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			jsonBody, _ := json.Marshal(testCase.ResponseBody)
			reqUrl, _ := url.Parse(testCase.RequestUrl)

			response := &http.Response{
				StatusCode: testCase.StatusCode,
				Body:       io.NopCloser(bytes.NewReader(jsonBody)),
				Request: &http.Request{
					Method: testCase.RequestMethod,
					URL:    reqUrl,
				},
			}

			// Act
			err := CheckResponse(response, testCase.ErrorResponse)

			// Asserts
			if testCase.ExpectedError && err == nil {
				t.Error("expected err got nil")
			}

			if testCase.ExpectedError && !reflect.DeepEqual(testCase.ErrorResponse, &testCase.ResponseBody) {
				t.Errorf("expected error response: %+v, got %+v", &testCase.ResponseBody, testCase.ErrorResponse)
			}
		})
	}
}

func TestDoMethod(t *testing.T) {
	testCases := []struct {
		Name          string
		ResponseBody  responseBody
		ExpectedError bool
		Method        string
		Path          string
	}{
		{
			Name: "Should return error for unsuccessful status code",
			ResponseBody: responseBody{
				Data: []data{
					{
						Field:  "Test field",
						Detail: "Error",
					},
				},
			},
			ExpectedError: true,
			Method:        "GET",
			Path:          "http://127.0.0.1:3001/error-test",
		},
		{
			Name: "Should not return error for successful status code",
			ResponseBody: responseBody{
				Data: []data{
					{
						Field:  "Test field",
						Detail: "Success",
					},
				},
			},
			ExpectedError: false,
			Method:        "GET",
			Path:          "http://127.0.0.1:3001/success-test",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			c, _ := NewClient()

			reqUrl, err := url.Parse(testCase.Path)
			if err != nil {
				t.Fatalf("error converting string to URL: %v", err)
			}

			request := &http.Request{
				Method: testCase.Method,
				URL:    reqUrl,
			}

			successResponse, errorResponse := new(responseBody), new(responseBody)

			// Act
			response, err := c.Do(request, successResponse, errorResponse)

			// Asserts
			if err != nil && response == nil {
				t.Fatalf("unable to execute request: %v", err)
			}

			if err != nil && !reflect.DeepEqual(errorResponse, &testCase.ResponseBody) {
				t.Errorf("expected error response: %+v, got %+v", &testCase.ResponseBody, errorResponse)
			}
		})
	}
}
