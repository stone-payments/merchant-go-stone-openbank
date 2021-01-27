package types

type TransferInput struct {
	AccountID   string `json:"account_id"`
	Amount      int    `json:"amount,omitempty"`
	Description string `json:"description,omitempty"`
	ScheduledTo string `json:"scheduled_to,omitempty"`
	Target      Target `json:"target,omitempty"`
	Type        string
}

type Transfer struct {
	ID                       string `json:"id,omitempty"`
	Amount                   int    `json:"amount,omitempty"`
	Fee                      int    `json:"fee,omitempty"`
	Target                   Target `json:"target,omitempty"`
	ApprovedAt               string `json:"approved_at,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	RejectedAt               string `json:"rejected_at,omitempty"`
	FailedAt                 string `json:"failed_at,omitempty"`
	FailureReasonCode        string `json:"failure_reason_code,omitempty"`
	FailureReasonDescription string `json:"failure_reason_description,omitempty"`
	Status                   string `json:"status,omitempty"`
	Description              string `json:"description,omitempty"`
	ApprovedBy               string `json:"approved_by,omitempty"`
	CreatedBy                string `json:"created_by,omitempty"`
	RejectedBy               string `json:"rejected_by,omitempty"`
	ApprovalExpiredAt        string `json:"approval_expired_at,omitempty"`
	CancelledAt              string `json:"cancelled_at,omitempty"`
	FinishedAt               string `json:"finished_at,omitempty"`
	ScheduledTo              string `json:"scheduled_to,omitempty"`

	RefundedAt               string
	RefundReasonCode         string `json:"refund_reason_code,omitempty"`
	RefundReasonDescription  string `json:"refund_reason_description,omitempty"`
	DelayedToNextBusinessDay bool   `json:"delayed_to_next_business_day,omitempty"`
	ScheduledToEffective     string `json:"scheduled_to_effective,omitempty"`
	ScheduledToRequested     string `json:"scheduled_to_requested,omitempty"`
}

type Target struct {
	Account TransferAccount `json:"account"`
	Entity  Entity          `json:"entity"`
}

type TransferAccount struct {
	AccountCode           string `json:"account_code,omitempty"`
	AccountType           string `json:"account_type,omitempty"`
	BranchCode            string `json:"branch_code,omitempty"`
	InstitutionISPB       string `json:"institution_ispb"`
	InstitutionCode       string `json:"institution_code"`
	InstitutionName       string `json:"institution_name,omitempty"`
	InstitutionNumberCode string `json:"institution_number_code"`
}

type Entity struct {
	Name         string `json:"name,omitempty"`
	Document     string `json:"document,omitempty"`
	DocumentType string `json:"document_type,omitempty"`
}
