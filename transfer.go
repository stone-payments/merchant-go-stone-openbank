package openbank

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/stone-co/go-stone-openbank/types"
)

// TransferService handles communication with Stone Openbank API
type TransferService struct {
	client *Client
}

func (s *TransferService) DryRunTransfer(input types.TransferInput, idempotencyKey string) (*types.Transfer, *Response, error) {
	path := "/api/v1/dry_run"
	return s.transfer(input, idempotencyKey, path)
}

func (s *TransferService) Transfer(input types.TransferInput, idempotencyKey string) (*types.Transfer, *Response, error) {
	path := "/api/v1"
	return s.transfer(input, idempotencyKey, path)
}

func (s *TransferService) transfer(input types.TransferInput, idempotencyKey, path string) (*types.Transfer, *Response, error) {
	var externalTransfer bool

	if input.Amount == 0 {
		return nil, nil, errors.New("amount can't be 0")
	}
	if input.AccountID == "" {
		return nil, nil, errors.New("account_id can't be empty")
	}
	if input.Target.Account.AccountCode == "" {
		return nil, nil, errors.New("account_code can't be empty")
	}

	if input.Target.Account.InstitutionCode != "" {
		if input.Target.Account.BranchCode == "" {
			return nil, nil, errors.New("branch_code can't be empty")
		}
		if input.Target.Entity.Name == "" {
			return nil, nil, errors.New("entity name can't be empty")
		}
		if input.Target.Entity.Document == "" {
			return nil, nil, errors.New("entity document can't be empty")
		}
		if input.Target.Entity.DocumentType == "" {
			return nil, nil, errors.New("entity document type can't be empty")
		}
		externalTransfer = true
	}

	if externalTransfer {
		path = fmt.Sprintf("%s/external_transfers", path)
	} else {
		path = fmt.Sprintf("%s/internal_transfers", path)
	}

	req, err := s.client.NewAPIRequest(http.MethodPost, path, input)
	if err != nil {
		return nil, nil, err
	}

	if idempotencyKey != "" {
		req.Header.Add("x-stone-idempotency-key", idempotencyKey)
	}

	var transfer types.Transfer
	resp, err := s.client.Do(req, &transfer)
	if err != nil {
		return nil, resp, err
	}

	return &transfer, resp, err
}
