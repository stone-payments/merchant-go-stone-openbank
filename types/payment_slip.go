package types

type PaymentSlipInput struct {
	AccountID      string                 `json:"account_id"`
	Amount         int                    `json:"amount"`
	ExpirationDate string                 `json:"expiration_date"`
	LimitDate      string                 `json:"limit_date,omitempty"`
	InvoiceType    string                 `json:"invoice_type"`
	Payer          *PaymentSlipPayerInput `json:"payer,omitempty"`
}

type PaymentSlipPayerInput struct {
	Document  string `json:"document"`
	LegalName string `json:"legal_name"`
	TradeName string `json:"trade_name,omitempty"`
}

type PaymentSlip struct {
	ID             string                  `json:"id"`
	AccountID      string                  `json:"account_id"`
	CreatedBy      string                  `json:"created_by"`
	CreatedAt      string                  `json:"created_at"`
	RegisteredAt   string                  `json:"registered_at"`
	SettledAt      string                  `json:"settled_at"`
	Amount         int                     `json:"amount"`
	Barcode        string                  `json:"barcode"`
	WritableLine   string                  `json:"writable_line"`
	ExpirationDate string                  `json:"expiration_date"`
	InvoiceType    string                  `json:"invoice_type"`
	IssuanceDate   string                  `json:"issuance_date"`
	LimitDate      string                  `json:"limit_date"`
	Status         string                  `json:"status"`
	OurNumber      string                  `json:"our_number"`
	Beneficiary    *PaymentSlipBeneficiary `json:"beneficiary"`
	Payer          *PaymentSlipPayer       `json:"payer"`
}

type PaymentSlipBeneficiary struct {
	AccountCode  string `json:"account_code"`
	BranchCode   string `json:"branch_code"`
	Document     string `json:"document"`
	DocumentType string `json:"document_type"`
	LegalName    string `json:"legal_name"`
	TradeName    string `json:"trade_name"`
}

type PaymentSlipPayer struct {
	Document     string `json:"document"`
	DocumentType string `json:"document_type"`
	LegalName    string `json:"legal_name"`
	TradeName    string `json:"trade_name"`
}
