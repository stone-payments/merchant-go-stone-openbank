package types

type Cursor struct {
	After  *int `json:"after"`
	Before *int `json:"before"`
	Limit  *int `json:"limit"`
}
