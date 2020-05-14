package openbank

import (
	"errors"
	"fmt"
	"github.com/Nhanderu/brdoc"
	"github.com/stone-co/go-stone-openbank/types"
	"regexp"
	"strings"
	"time"
)

const (
	invoiceAmountMin = 2000
	invoiceAmountMax = 1000000

	invoiceTypeDeposit        = "deposit"
	invoiceTypeProposal       = "proposal"
	invoiceTypeBillOfExchange = "bill_of_exchange"
)

var digitsRegex = regexp.MustCompile("[0-9]+")

func onlyDigits(key string) string {
	return strings.Join(digitsRegex.FindAllString(key, -1), "")
}

// PaymentInvoiceService handles communication with Stone Openbank API
type PaymentSlipService struct {
	client *Client
}

// PaymentSlip make
func (s *PaymentSlipService) PaymentSlip(input types.PaymentInvoiceInput, idempotencyKey string) (*types.PaymentInvoice, *Response, error) {
	path := "/api/v1/barcode_payment_invoices"
	return s.paymentSlip(input, idempotencyKey, path)
}

func (s *PaymentSlipService) paymentSlip(input types.PaymentInvoiceInput, idempotencyKey string, path string) (*types.PaymentInvoice, *Response, error) {

	if strings.TrimSpace(input.AccountID) == "" {
		return nil, nil, errors.New("account_id can't be empty")
	}

	if input.Amount < invoiceAmountMin || input.Amount > invoiceAmountMax {
		return nil, nil, fmt.Errorf("amount can't be < %v or > %v", invoiceAmountMin, invoiceAmountMax)
	}

	_, err := time.Parse("2006-01-02", input.ExpirationDate)
	if err != nil || time.Now().Format("2006-01-02") < input.ExpirationDate {
		return nil, nil, errors.New("invalid expiration_date")
	}

	switch input.InvoiceType {
	case invoiceTypeDeposit, invoiceTypeProposal:
		input.LimitDate = input.ExpirationDate
	case invoiceTypeBillOfExchange:
		if strings.TrimSpace(input.LimitDate) == "" {
			input.LimitDate = input.ExpirationDate
		} else {
			_, err := time.Parse("2006-01-02", input.LimitDate)
			if err != nil || input.LimitDate < input.ExpirationDate {
				return nil, nil, errors.New("invalid limit_date")
			}
		}
	default:
		return nil, nil, errors.New("invalid invoice_type")
	}

	if input.InvoiceType != invoiceTypeDeposit {
		if strings.TrimSpace(input.Payer.LegalName) == "" {
			return nil, nil, errors.New("invalid payer legal_name")
		}

		input.Payer.Document = onlyDigits(input.Payer.Document)
		if !brdoc.IsCPF(input.Payer.Document) && !brdoc.IsCNPJ(input.Payer.Document) {

		}
	}





	return nil, nil, nil
}
