package static

var (
	// return payment pay
	StaticURL   = "https://support.bk.cloudns.cc"
	okpage      = "/payment/static/return" // fixed back to static return page
	forwardpage = "/pay.html"
)

type Client struct {
	Method string `json:"method"`
}

func (c *Client) GetReturnPage() string {
	return okpage
}

func (c *Client) GetPaymentPage() string {
	return forwardpage
}
