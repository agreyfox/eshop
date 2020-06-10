package payssion

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// NewClient returns new Client struct
// APIBase is a base API URL, for testing you can use paypal.APIBaseSandBox
func NewClient(apikey string, secret string, APIBase string) (*Client, error) {
	if apikey == "" || secret == "" || APIBase == "" {
		return nil, errors.New("ClientID, Secret and APIBase are required to create a Payssion Client")
	}
	return &Client{
		Client:   &http.Client{},
		ClientID: apikey,
		Secret:   secret,
		APIBase:  APIBase,
		Log:      logger,
	}, nil
}

// SetLog will set/change the output destination.
// If log file is set paypal will log all requests and responses to this Writer
func (c *Client) SetLog(log *zap.SugaredLogger) {
	c.Log = log
}

// Send makes a request to the API, the response body will be
// unmarshaled into v, or if v is an io.Writer, the response will
// be written to it without decoding
func (c *Client) Send(req *http.Request, v interface{}) error {
	var (
		err  error
		resp *http.Response
		data []byte
	)

	// Set default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en_US")

	// Default values for headers
	if req.Header.Get("Content-type") == "" {
		req.Header.Set("Content-type", "application/json")
	}

	resp, err = c.Client.Do(req)
	c.log(req, resp)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errResp := &ErrorResponse{Response: resp}
		data, err = ioutil.ReadAll(resp.Body)

		if err == nil && len(data) > 0 {
			json.Unmarshal(data, errResp)
		}

		return errResp
	}
	if v == nil {
		return nil
	}

	if w, ok := v.(io.Writer); ok {
		io.Copy(w, resp.Body)
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// SendWithAuth makes a request to the API and apply OAuth2 header automatically.
// If the access token soon to be expired or already expired, it will try to get a new one before
// making the main request
// client.Token will be updated when changed
func (c *Client) SendWithAuth(req *http.Request, v interface{}) error {
	c.Lock()

	// Unlock the client mutex before sending the request, this allows multiple requests
	// to be in progress at the same time.
	c.Unlock()
	return c.Send(req, v)
}

// NewRequest constructs a request
// Convert payload to a JSON
func (c *Client) NewRequest(method, url string, payload interface{}) (*http.Request, error) {
	var buf io.Reader
	if payload != nil {
		b, err := json.Marshal(&payload)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	}

	return http.NewRequest(method, url, buf)
}

// send form data
func (c *Client) SendWithData(url string, v url.Values) (*http.Response, error) {

	return http.PostForm(url, v)

}

// log will dump request and response to the log file
func (c *Client) log(r *http.Request, resp *http.Response) {
	if c.Log != nil {
		var (
			reqDump  string
			respDump []byte
		)

		if r != nil {
			reqDump = fmt.Sprintf("%s %s. Data: %s", r.Method, r.URL.String(), r.Form.Encode())
		}
		if resp != nil {
			respDump, _ = httputil.DumpResponse(resp, true)
		}

		c.Log.Debugf(fmt.Sprintf("Request: %s\nResponse: %s\n", reqDump, string(respDump)))
	}
}

// create signature from order information
func Signature(ar []string) string {
	ma := strings.Join(ar, "|")
	hash := md5.Sum([]byte(ma))
	//fmt.Println(ma, hash, hex.EncodeToString(hash[:]))
	return hex.EncodeToString(hash[:])

}

func isRightPMID(pm_id string) bool {
	found := false
	for _, item := range thirdParty {
		if pm_id == item.PMID {
			found = true
			break
		}
	}
	return found
}

// to create order
func (c *Client) CreateOrder(pr *PaymentRequest) (*PaymentResponse, error) {
	re := &PaymentResponse{}

	v, err := json.Marshal(pr)
	if err != nil {
		logger.Error(err)
		return re, err
	}
	reqContent := bytes.NewBuffer(v)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s%s", c.APIBase, "/api/v1/payment/create"), reqContent)

	req.Header.Set("Content-type", "application/json")
	response, _ := c.Client.Do(req)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(body))

		if err != nil {
			logger.Error("Error:", err.Error())
			return re, err
		}
		//rr := map[string]interface{}{}
		err = json.Unmarshal(body, re)
		if err != nil {
			logger.Error("Error:", err.Error())
			return re, err
		}
		return re, nil

	}
	return re, errors.New("Retrun error value")

}

func (c *Client) GetOrder(tr, orderid string) {

}
