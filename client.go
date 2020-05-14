package openbank

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
)

const (
	libraryVersion    = "1.0"
	prodAccountURL    = "https://accounts.openbank.stone.com.br"
	sandboxAccountURL = "https://sandbox-accounts.openbank.stone.com.br"
	prodAPIBaseURL    = "https://api.openbank.stone.com.br"
	sandboxAPIBaseURL = "https://sandbox-api.openbank.stone.com.br"
	userAgent         = "go-stone-openbank/" + libraryVersion
)

type Client struct {
	client *http.Client
	log    *logrus.Entry
	debug  bool

	AccountURL *url.URL
	ApiBaseURL *url.URL

	ClientID           string
	ConsentRedirectURL string
	PrivateKeyPath     string

	Sandbox bool

	UserAgent string

	Token string

	//Services used for comunicating with API
	Account  *AccountService
	Transfer *TransferService
	PaymentInvoice *PaymentInvoiceService
}

//vhttpClient *http.Client, sandbox bool, clientID, consentRedirectURL strinxg
func NewClient(opts ...ClientOpt) *Client {

	accountURL, _ := url.Parse(prodAccountURL)
	apiURL, _ := url.Parse(prodAPIBaseURL)

	c := Client{
		client:     http.DefaultClient,
		UserAgent:  userAgent,
		AccountURL: accountURL,
		ApiBaseURL: apiURL,
	}

	c.ApplyOpts(opts...)

	log := logrus.New().WithFields(logrus.Fields{
		"apiURL":     c.ApiBaseURL,
		"accountURL": c.AccountURL,
	})

	c.log = log

	//Set services
	c.Account = &AccountService{client: &c}
	c.Transfer = &TransferService{client: &c}

	return &c
}

type ClientOpt func(*Client)

func WithClientID(key string) ClientOpt {
	return func(c *Client) {
		c.ClientID = key
	}
}

func SetPrivateKey(path string) ClientOpt {
	return func(c *Client) {
		c.PrivateKeyPath = path
	}
}

func SetConsentURL(url string) ClientOpt {
	return func(c *Client) {
		c.ConsentRedirectURL = url
	}
}

func WithHttpClient(hc http.Client) ClientOpt {
	return func(c *Client) {
		c.client = &hc
	}
}

func UseSandbox() ClientOpt {
	return func(c *Client) {
		accountURL, _ := url.Parse(sandboxAccountURL)
		apiURL, _ := url.Parse(sandboxAPIBaseURL)
		c.Sandbox = true
		c.ApiBaseURL = apiURL
		c.AccountURL = accountURL
	}
}

func SetUserAgent(ua string) ClientOpt {
	return func(c *Client) {
		c.UserAgent = fmt.Sprintf("%s %s", ua, c.UserAgent)
	}
}

func EnableDebug() ClientOpt {
	return func(c *Client) {
		c.debug = true
	}

}

func (c *Client) ApplyOpts(opts ...ClientOpt) {
	for _, opt := range opts {
		opt(c)
	}
}

type Response struct {
	*http.Response
}

type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`

	// RequestID returned from the API, useful to contact support.
	RequestID string `json:"request_id"`
}

func (r *ErrorResponse) Error() string {
	if r.RequestID != "" {
		return fmt.Sprintf("%v %v: %d (request %q) %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}

	return errorResponse
}

// NewAPIRequest creates an API request. A relative URL PATH can be provided in pathStr, which will be resolved to the
// ApiBaseURL of the Client.
func (c *Client) NewAPIRequest(method, pathStr string, body interface{}) (*http.Request, error) {
	u, err := c.ApiBaseURL.Parse(pathStr)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	if c.debug {
		d, _ := httputil.DumpRequestOut(req, true)
		c.log.Infof(">>> REQUEST:\n%s", string(d))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()
	if c.debug {
		dr, _ := httputil.DumpResponse(resp, true)
		c.log.Infof("<<< RESULT:\n%s", string(dr))
	}

	response := &Response{Response: resp}

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, err
}
