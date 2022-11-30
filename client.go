package openbank

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/stone-co/go-stone-openbank/types"
)

const (
	libraryVersion        = "1.0"
	prodAccountURL        = "https://accounts.openbank.stone.com.br"
	sandboxAccountURL     = "https://sandbox-accounts.openbank.stone.com.br"
	prodAPIBaseURL        = "https://api.openbank.stone.com.br"
	sandboxAPIBaseURL     = "https://sandbox-api.openbank.stone.com.br"
	prodSiteURL           = "https://conta.stone.com.br"
	sandboxSiteURL        = "https://sandbox.conta.stone.com.br"
	userAgent             = "go-stone-openbank/" + libraryVersion
	idempotencyKeyMaxSize = 72
)

type Client struct {
	client *http.Client
	log    *logrus.Entry
	m      *sync.Mutex
	debug  bool

	AccountURL *url.URL
	ApiBaseURL *url.URL
	SiteURL    *url.URL

	StonePublicKeys types.StonePublicKeys

	ClientID           string
	ConsentRedirectURL string

	privateKeyData []byte // used to build privateKey
	privateKey     *rsa.PrivateKey

	Sandbox bool

	UserAgent string

	token oauth2.Token

	//Services used for comunicating with API
	Institution    *InstitutionService
	Account        *AccountService
	Transfer       *TransferService
	PaymentInvoice *PaymentInvoiceService
	Pix            *PixService
	PaymentLink    *PaymentLinkService
	Topups         *TopupsService
}

func NewClient(opts ...ClientOpt) (*Client, error) {
	accountURL, _ := url.Parse(prodAccountURL)
	apiURL, _ := url.Parse(prodAPIBaseURL)
	siteURL, _ := url.Parse(prodSiteURL)

	c := Client{
		client:          http.DefaultClient,
		UserAgent:       userAgent,
		AccountURL:      accountURL,
		ApiBaseURL:      apiURL,
		SiteURL:         siteURL,
		StonePublicKeys: make(types.StonePublicKeys),
		m:               &sync.Mutex{},
	}

	c.ApplyOpts(opts...)

	if len(c.privateKeyData) > 0 {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(c.privateKeyData)
		if err != nil {
			return nil, err
		}

		if privateKey == nil {
			return nil, fmt.Errorf("invalid private key")
		}

		c.privateKey = privateKey
	}

	//Set services
	c.Account = &AccountService{client: &c}
	c.Institution = &InstitutionService{client: &c}
	c.PaymentLink = &PaymentLinkService{client: &c}
	c.PaymentInvoice = &PaymentInvoiceService{client: &c}
	c.Pix = &PixService{client: &c}
	c.Topups = &TopupsService{client: &c}
	c.Transfer = &TransferService{client: &c}

	// Set log
	log := logrus.New().WithFields(logrus.Fields{
		"apiURL":     c.ApiBaseURL.String(),
		"accountURL": c.AccountURL.String(),
		"siteURL":    c.SiteURL.String(),
	})
	c.log = log

	return &c, nil
}

type ClientOpt func(*Client)

func WithClientID(key string) ClientOpt {
	return func(c *Client) {
		c.ClientID = key
	}
}

func WithPEMPrivateKey(pk []byte) ClientOpt {
	return func(c *Client) {
		c.privateKeyData = pk
	}
}

func SetConsentURL(url string) ClientOpt {
	return func(c *Client) {
		c.ConsentRedirectURL = url
	}
}

func SetBaseURL(newBaseUrl string) (ClientOpt, error) {
	apiBaseURL, err := url.Parse(newBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	return func(c *Client) {
		c.ApiBaseURL = apiBaseURL
	}, nil
}

func SetAccountURL(newAccountUrl string) (ClientOpt, error) {
	accountURL, err := url.Parse(newAccountUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid account url: %w", err)
	}
	return func(c *Client) {
		c.AccountURL = accountURL
	}, nil
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
		siteURL, _ := url.Parse(sandboxSiteURL)
		c.Sandbox = true
		c.ApiBaseURL = apiURL
		c.AccountURL = accountURL
		c.SiteURL = siteURL
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
	if opts == nil {
		return
	}
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

	TransferError TransferError `json:"transfer_error"`
}

type TransferError struct {
	Type             string `json:"type,omitempty"`
	ValidationErrors []struct {
		Error string   `json:"error,omitempty"`
		Path  []string `json:"path,omitempty"`
	} `json:"validation_errors,omitempty"`
	Reason []struct {
		Error string   `json:"error,omitempty"`
		Path  []string `json:"path,omitempty"`
	} `json:"reason,omitempty"`
}

func (r *ErrorResponse) Error() string {
	if r.RequestID != "" {
		return fmt.Sprintf("%v %v: %d (request %q) %v %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.TransferError, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %v %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.TransferError, r.Message)
}

func CheckResponse(r *http.Response, responseBody interface{}) (bool, error) {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return true, nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		if responseBody != nil {
			if w, ok := responseBody.(io.Writer); ok {
				_, err = io.Copy(w, r.Body)
				if err != nil {
					return false, err
				}
			} else {
				err = json.Unmarshal(data, responseBody)
				if err != nil {
					return false, err
				}

				return false, nil
			}
		}

		err := json.Unmarshal(data, &errorResponse.TransferError)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}

	return false, errorResponse
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

//AddIdempotencyHeader add in request the header used to realize idempotent operations
func (c *Client) AddIdempotencyHeader(req *http.Request, idempotencyKey string) error {
	trimmedIdempotencyKey := strings.TrimSpace(idempotencyKey)
	if trimmedIdempotencyKey != "" {
		if len(trimmedIdempotencyKey) > idempotencyKeyMaxSize {
			return errors.New("invalid idempotency key")
		}
		req.Header.Add("x-stone-idempotency-key", trimmedIdempotencyKey)
	}

	return nil
}

//AddAccountIdHeader add in request the header used in some pix operations and maybe others
func (c *Client) AddAccountIdHeader(req *http.Request, accountId string) error {
	trimmedAccountId := strings.TrimSpace(accountId)
	if trimmedAccountId != "" {
		req.Header.Add("x-stone-account-id", trimmedAccountId)
	}

	return nil
}

func (c *Client) Do(req *http.Request, successResponse, errorResponse interface{}) (*Response, error) {
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

	ok, err := CheckResponse(resp, errorResponse)
	if err != nil {
		return response, err
	}

	if ok && successResponse != nil {
		if w, ok := successResponse.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(successResponse)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, err
}
