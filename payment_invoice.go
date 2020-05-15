package openbank

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/stone-co/go-stone-openbank/types"
)

const idempotencyKeyMaxSize = 72

// PaymentInvoiceService handlers communication with Stone Openbank API
type PaymentInvoiceService struct {
	client *Client
}

// PaymentInvoice make a bar code payment invoice
func (s *PaymentInvoiceService) PaymentInvoice(input types.PaymentInvoiceInput, idempotencyKey string) (*types.PaymentInvoice, *Response, error) {
	path := "/api/v1/barcode_payment_invoices"
	if err := input.Ok(); err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewAPIRequest(http.MethodPost, path, input)
	if err != nil {
		return nil, nil, err
	}

	if idempotencyKey != "" {
		if len(idempotencyKey) > idempotencyKeyMaxSize {
			return nil, nil, errors.New("invalid idempotency key")
		}
		req.Header.Add("x-stone-idempotency-key", idempotencyKey)
	}

	var paymentInvoice types.PaymentInvoice
	resp, err := s.client.Do(req, &paymentInvoice)
	if err != nil {
		return nil, resp, err
	}

	return &paymentInvoice, resp, err
}

// List returns a list of PaymentInvoices
func (s *PaymentInvoiceService) List(accountID string) ([]types.PaymentInvoice, *Response, error) {
	path := fmt.Sprintf("/api/v1/barcode_payment_invoices/?account_id=%s", accountID)
	if strings.TrimSpace(accountID) == "" {
		return nil, nil, errors.New("account_id can't be empty")
	}

	req, err := s.client.NewAPIRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var dataResp struct {
		Cursor types.Cursor     `json:"cursor"`
		Data   []types.PaymentInvoice `json:"data"`
	}

	resp, err := s.client.Do(req, &dataResp)
	if err != nil {
		return nil, resp, err
	}

	return dataResp.Data, resp, err
}

// Get return a PaymentInvoice
func (s *PaymentInvoiceService) Get(paymentInvoiceID string) (types.PaymentInvoice, *Response, error) {
	path := fmt.Sprintf("/api/v1/barcode_payment_invoices/%s", paymentInvoiceID)
	var paymentInvoice types.PaymentInvoice
	if strings.TrimSpace(paymentInvoiceID) == "" {
		return paymentInvoice, nil, errors.New("payment_invoice_id can't be empty")
	}

	req, err := s.client.NewAPIRequest(http.MethodGet, path, nil)
	if err != nil {
		return paymentInvoice, nil, err
	}

	resp, err := s.client.Do(req, &paymentInvoice)
	if err != nil {
		return paymentInvoice, resp, err
	}

	return paymentInvoice, resp, err
}