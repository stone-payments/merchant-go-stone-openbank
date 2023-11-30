package openbank

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"github.com/stone-payments/merchant-go-stone-openbank/types"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
)

const (
	libraryVersion        = "1.0"
	prodAccountURL        = "https://accounts.openbank.stone.com.br"
	sandboxAccountURL     = "https://sandbox-accounts.openbank.stone.com.br"
	prodAPIBaseURL        = "https://api.openbank.stone.com.br"
	sandboxAPIBaseURL     = "https://sandbox-api.openbank.stone.com.br"
	prodSiteURL           = "https://conta.stone.com.br"
	sandboxSiteURL        = "https://sandbox.conta.stone.com.br"
	userAgent             = "merchant-go-stone-openbank/" + libraryVersion
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

	otelTracer trace.Tracer
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

func WithTracer(tracer trace.Tracer) ClientOpt {
	return func(c *Client) {
		c.otelTracer = tracer
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

	TransferError interface{} `json:"transfer_error"`
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
		return fmt.Sprintf("%v %v: %d (request %q) %+v %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.TransferError, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %+v %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.TransferError, r.Message)
}

func parseBody(data []byte, body interface{}) error {
	if len(data) > 0 && body != nil {
		if w, ok := body.(io.Writer); ok {
			_, err := io.Copy(w, bytes.NewReader(data))
			if err != nil {
				return err
			}
		} else {
			err := json.Unmarshal(data, body)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func CheckResponse(r *http.Response, errorBody interface{}) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil {
		if err = parseBody(data, errorBody); err != nil {
			errorResponse.Message = string(data)
		}
		errorResponse.TransferError = errorBody
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

// AddIdempotencyHeader add in request the header used to realize idempotent operations
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

func (c *Client) Do(req *http.Request, successResponse, errorResponse interface{}) (*Response, error) {
	_, span := c.newSpan(
		req.Context(),
		"merchant openbank client request",
		trace.SpanKindClient,
	)
	defer c.endSpan(span)

	c.addSpanAttribute(span, attribute.String("http.request.path", req.URL.String()))
	c.addSpanAttribute(span, attribute.String("http.request.protocol", req.Proto))
	c.addSpanAttribute(span, attribute.String("http.request.method", req.Method))

	if c.debug {
		d, _ := httputil.DumpRequestOut(req, true)
		c.log.Infof(">>> REQUEST:\n%s", string(d))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.setSpanStatus(span, codes.Error, "error executing request")
		c.recordError(span, err)
		return nil, err
	}

	c.addSpanAttribute(span, attribute.Int("http.response.status_code", resp.StatusCode))

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

	err = CheckResponse(resp, errorResponse)
	if err != nil {
		c.setSpanStatus(span, codes.Error, "client request error")
		c.spanRecordError(span, err)

		return response, err
	}

	data, err := io.ReadAll(resp.Body)
	if err = parseBody(data, successResponse); err != nil {
		c.setSpanStatus(span, codes.Error, "client request error")
		c.spanRecordError(span, err)

		return response, err
	}

	c.setSpanStatus(span, codes.Ok, "client request succeeded")

	return response, err
}

func (c *Client) newSpan(ctx context.Context, name string, kind trace.SpanKind) (context.Context, trace.Span) {
	if c.otelTracer != nil {
		return c.otelTracer.Start(ctx, name, trace.WithSpanKind(kind))
	}

	return ctx, nil
}

func (c *Client) setSpanStatus(span trace.Span, code codes.Code, description string) {
	if span != nil {
		span.SetStatus(code, description)
	}
}

func (c *Client) spanRecordError(span trace.Span, err error) {
	if span != nil {
		span.RecordError(err)
	}
}

func (c *Client) endSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}

func (c *Client) addSpanAttribute(span trace.Span, attributes ...attribute.KeyValue) {
	if span != nil {
		span.SetAttributes(attributes...)
	}
}

func (c *Client) recordError(span trace.Span, err error) {
	if span != nil {
		span.RecordError(err)
	}
}
