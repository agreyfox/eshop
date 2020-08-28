package paypal

import (
	"fmt"
	"strconv"
	"time"
)

type TransactionSearchRequest struct {
	TransactionID               *string `json:"transactions_id,omitempty"`
	TransactionType             *string
	TransactionStatus           *string
	TransactionAmount           *string
	TransactionCurrency         *string
	StartDate                   time.Time
	EndDate                     time.Time
	PaymentInstrumentType       *string
	StoreID                     *string
	TerminalID                  *string
	Fields                      *string
	BalanceAffectingRecordsOnly *string
	PageSize                    *int
	Page                        *int
}

type TransactionSearchResponse struct {
	TransactionDetails  []SearchTransactionDetails `json:"transaction_details"`
	AccountNumber       string                     `json:"account_number"`
	StartDate           string                     `json:"start_date"`
	EndDate             string                     `json:"end_date"`
	LastRefreshDatetime string                     `json:"last_refreshed_datetime"`
	Page                int                        `json:"page"`
	TotalItem           int                        `json:"total_item,omitempty"`
	TotalPage           int                        `json:"total_page,omitempty"`
	Links               []Link                     `json:"links,omitempty"`
}

// ListTransactions - Use this to search PayPal transactions from the last 31 days.
// Endpoint: GET /v1/reporting/transactions
func (c *Client) ListTransactions(req *TransactionSearchRequest) (*TransactionSearchResponse, error) {
	response := &TransactionSearchResponse{}

	r, err := c.NewRequest("GET", fmt.Sprintf("%s%s", c.APIBase, "/v1/reporting/transactions"), nil)
	if err != nil {
		return nil, err
	}

	q := r.URL.Query()

	q.Add("start_date", req.StartDate.Format(time.RFC3339))
	q.Add("end_date", req.EndDate.Format(time.RFC3339))

	if req.TransactionID != nil {
		q.Add("transactions_id", *req.TransactionID)
	}
	if req.TransactionType != nil {
		q.Add("transaction_type", *req.TransactionType)
	}
	if req.TransactionStatus != nil {
		q.Add("transaction_status", *req.TransactionStatus)
	}
	if req.TransactionAmount != nil {
		q.Add("transaction_amount", *req.TransactionAmount)
	}
	if req.TransactionCurrency != nil {
		q.Add("transaction_currency", *req.TransactionCurrency)
	}
	if req.PaymentInstrumentType != nil {
		q.Add("payment_instrument_type", *req.PaymentInstrumentType)
	}
	if req.StoreID != nil {
		q.Add("store_id", *req.StoreID)
	}
	if req.TerminalID != nil {
		q.Add("terminal_id", *req.TerminalID)
	}
	if req.Fields != nil {
		q.Add("fields", *req.Fields)
	}
	if req.BalanceAffectingRecordsOnly != nil {
		q.Add("balance_affecting_records_only", *req.BalanceAffectingRecordsOnly)
	}
	if req.PageSize != nil {
		q.Add("page_size", strconv.Itoa(*req.PageSize))
	}
	if req.Page != nil {
		q.Add("page", strconv.Itoa(*req.Page))
	}

	r.URL.RawQuery = q.Encode()
	/* var dd string

	buf := bytes.NewBufferString(dd) */
	logger.Debug("Url is ", r.URL)
	if err = c.SendWithAuth(r, response); err != nil {
		return nil, err
	}

	return response, nil
}
