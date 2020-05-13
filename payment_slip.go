package openbank

import (
	"errors"
	"fmt"
	"github.com/stone-co/go-stone-openbank/types"
	"strings"
	"time"
)

const (
	paymentSlipAmountMin = 2000
	paymentSlipAmountMax = 1000000

	paymentSlipTypeDeposit        = "deposit"
	paymentSlipTypeProposal       = "proposal"
	paymentSlipTypeBillOfExchange = "bill_of_exchange"
)

// PaymentSlipService handles communication with Stone Openbank API
type PaymentSlipService struct {
	client *Client
}

// PaymentSlip make
func (s *PaymentSlipService) PaymentSlip(input types.PaymentSlipInput, idempotencyKey string) (*types.PaymentSlip, *Response, error) {
	path := "/api/v1/barcode_payment_invoices"
	return s.paymentSlip(input, idempotencyKey, path)
}

func (s *PaymentSlipService) paymentSlip(input types.PaymentSlipInput, idempotencyKey string, path string) (*types.PaymentSlip, *Response, error) {

	if strings.TrimSpace(input.AccountID) == "" {
		return nil, nil, errors.New("account_id can't be empty")
	}

	if input.Amount < paymentSlipAmountMin || input.Amount > paymentSlipAmountMax {
		return nil, nil, fmt.Errorf("amount can't be < %v or > %v", paymentSlipAmountMin, paymentSlipAmountMax)
	}

	_, err := time.Parse("2006-01-02", input.ExpirationDate)
	if err != nil {
		return nil, nil, errors.New("invalid date")
	}

	return nil, nil, nil
}
