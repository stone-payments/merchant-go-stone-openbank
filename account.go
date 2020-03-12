package openbank

import (
	"fmt"
	"net/http"
)

// AccountService handles communication with Stone Openbank API
type AccountService struct {
	client *Client
}

//Account represents a Stone PaymentAccount
type Account struct {
	AccountCode        string `json:"account_code"`
	BranchCode         string `json:"branch_code"`
	ID                 string `json:"id"`
	OwnerDocument      string `json:"owner_document"`
	OwnerID            string `json:"owner_id"`
	OwnerName          string `json:"owner_name"`
	RestrictedFeatures bool   `json:"restricted_features"`
}

// Get account info
func (s *AccountService) Get(id string) (*Account, *Response, error) {

	path := fmt.Sprintf("/api/v1/accounts/%s", id)

	req, err := s.client.NewAPIRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var account Account
	resp, err := s.client.Do(req, &account)
	if err != nil {
		return nil, resp, err
	}

	return &account, resp, err
}

// List accounts
func (s *AccountService) List() ([]Account, *Response, error) {

	path := "/api/v1/accounts?paginate=true"

	req, err := s.client.NewAPIRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var dataResp struct {
		Cursor struct {
			After  *int
			Before *int
			Limit  *int
		} `json:"cursor"`
		Data []Account `json:"data"`
	}

	resp, err := s.client.Do(req, &dataResp)
	if err != nil {
		return nil, resp, err
	}

	return dataResp.Data, resp, err
}
