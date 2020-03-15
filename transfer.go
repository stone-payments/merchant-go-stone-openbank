package openbank

import (
	"errors"

	"net/http"

	"github.com/stone-co/go-stone-openbank/types"
)

// TransferService handles communication with Stone Openbank API
type TransferService struct {
	client *Client
}

func (s *TransferService) DryRunTransfer(input types.TransferInput, idempotencyKey string) (*types.Transfer, *Response, error) {
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

	path := "/api/v1/dry_run/internal_transfers"

	if externalTransfer {
		path = "/api/v1/dry_run/external_transfers"
	}
	req, err := s.client.NewAPIRequest(http.MethodPost, path, input)
	if err != nil {
		return nil, nil, err
	}
	var transfer types.Transfer
	resp, err := s.client.Do(req, &transfer)
	if err != nil {
		return nil, resp, err
	}

	return &transfer, resp, err
}

func (s *TransferService) Transfer(input types.TransferInput, idempotencyKey string) (*types.Transfer, *Response, error) {
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

	path := "/api/v1/internal_transfers"

	if externalTransfer {
		path = "/api/v1/external_transfers"
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
