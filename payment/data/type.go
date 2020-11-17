package data

import (
	"errors"

	"github.com/agreyfox/eshop/system/logs"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	IDMaker string = "--"

	DescriptionMarker string = "||"

	OrderCreated    = "已创建"
	OrderPaid       = "已付款"
	OrderInValidate = "待检验"
	OrderInDelivery = "待交付"
	OrderCompleted  = "已完成"
	OrderCancel     = "用户取消"
	OrderRefunded   = "已退款"
	OrderDisputed   = "用户欺诈"
)

var (
	PaymentErrOpenDB    = errors.New("数据库错误")
	PaymentErrInputData = errors.New("用户数据错误")
	PaymentErrSave      = errors.New("保存数据失败")
)

var (
	DBRequestFile = ".request.db"
	DBRequest     = "requests" // save the order request table name

	DBPaymentLog = ".logs.db" //save request to log
	DBLogName    = "logs"     //log table name

	UserRequest = "request"
	Complete    = "order"
	DbFile      = "records.db"

	OrderName = "Order" //in main system.db
)

var (
	logger *zap.SugaredLogger = logs.Log.Sugar()

	PaymentDBHandler  *bolt.DB //order request db
	PaymentLogHandler *bolt.DB // log db

	SystemDBHandler *bolt.DB //system db
	OnlineURL       = "https://support.bk.cloudns.cc/#/Result"
)

type (
	Item struct {
		Game      string  `json:"game"`
		Server    string  `json:"server,omitempty"`
		Category  string  `json:"category"`
		Product   string  `json:"product"`
		UnitPrice float64 `json:"unit_price"`
		Quantity  int     `json:"quantity"`
	}
	// User submit same struct to system for order creation
	UserSubmitOrderRequest struct {
		OrderID        string  `json:"order_id,omitempty"`   // 订单号
		OrderDate      int64   `json:"order_date,omitempty"` //订单日期 millionseconds
		Payment        string  `json:"payment"`
		PaymentChannel string  `json:"payment_channel"`
		Country        string  `json:"county,omitempty"`
		City           string  `json:"city,omitempty"`
		Currency       string  `json:"currency"`
		Language       string  `json:"language"`
		Locale         string  `json:"locale"`
		Email          string  `json:"email"`
		FirstName      string  `json:"first_name"`
		LastName       string  `json:"last_name"`
		IPAddr         string  `json:"ip"`
		ContactInfo    string  `json:"contact_info"`
		Phone          string  `json:"phone,omitempty"`
		Amount         float64 `json:"amount"`
		SubTotal       float64 `json:"sub_total"`
		ItemList       []Item  `json:"item_list"`
		RequestInfo    string  `json:"request_info,omitempty"`
		CouponCode     string  `json:"coupon_code,omitempty"`
		CouponValue    float64 `json:"coupon_value,omitempty"`
		PaymentFee     float64 `json:"payment_fee,omitempty"`
		LogoURL        string  `json:"logo_url,omitempty"`
		Address        string  `json:"address,omitempty"`
		Status         string  `json:"status"`
		Description    string  `json:"description"`
		Respond        string  `json:"respond,omitempty"`
	}

	//save data to usrerequest
	PaymentLog struct {
		PaymentMethod string      `json:"method"`
		PaymentID     string      `json:"payment_id"`
		PaymentState  string      `json:"payment_state"`
		OrderID       string      `json:"order_id"`
		Total         string      `json:"total"`
		Currency      string      `json:"currency"`
		RequestData   interface{} `json:"request,omitempty"`
		ReturnData    interface{} `json:"return,omitempty"`
		RequestTime   int64       `json:"request_time"`
		IP            string      `json:"ip,omitempty"`
		BuyerEmail    string      `json:"email,omitempty"`
		Info          string      `json:"info,omitempty"`
		Comments      string      `json:"comments,omitempty"`
		Address       string      `json:"delivery_address,omitempty"`
		Description   string      `json:"description,omitempty"`
	}

	PaymentRecord struct {
		Method        string      `json:"method"`
		PaymentID     string      `json:"payment_id"`
		PaymentAction string      `json:"payment_action,omitempty"`
		Request       interface{} `json:"request,omitempty"`
		Result        interface{} `json:"result,omitempty"`
		Currency      string      `json:"currency"`
		Total         string      `json:"total"`
		Fee           string      `json:"fee,omitempty"`
		Tax           string      `json:"tax,omitempty"`
		Discount      string      `json:"discount,omitempty"`
		Delivery      string      `json:"delivery,omitempty"`
		PaymentState  string      `json:"state"`
		Status        string      `json:"status"`
		BuyerEmail    string      `json:"email"`
		BuyerComments string      `json:"comment,omitempty"`
		Description   string      `json:"description,omitempty"`
		RequestTime   int64       `json:"request_time"`
		RequestIP     string      `json:"ip,omitempty"`
	}

	// Order struct in eshop Beckend system
	// the order struct ID:value, 这个id就是 transaction_id||order
	Order struct {
		//ID             string      `json:"id"`

		Status         string `json:"status"`
		OrderDetail    string `json:"order_detail,omitempty"`
		OrderID        string `json:"order_id"`
		PaymentID      string `json:"payment_id,omitempty"`
		TransactionID  string `json:"transaction_id,omitempty"`
		PaymentVendor  string `json:"vendor"`
		PaymentMethod  string `json:"method"`
		PaymentNote    string `json:"payment_note,omitempty"`
		Payer          string `json:"payer"`
		PayerLink      string `json:"payer_link"`
		PayerIP        string `json:"ip,omitempty"`
		Currency       string `json:"currency"`
		Total          string `json:"total"`
		Paid           string `json:"paid,omitempty"`
		Net            string `json:"net,omitempty"`
		Description    string `json:"description,omitempty"`
		NotifyInfo     string `json:"notify_info"`
		Paytime        string `json:"pay_time,omitempty"`
		DeliveryTime   string `json:"delivery_time,omitempty"`
		DeliveryUserID string `json:"worker,omitempty"`
		Comments       string `json:"comments,omitempty"`
		AdminNote      string `json:"note,omitempty"`
		IsRefund       bool   `json:"is_refund,omitempty"`
		IsChargeBack   bool   `json:"is_chargeback,omitempty"`
		RequestTime    string `json:"request_time,omitempty"`
		RefundTime     string `json:"refund_time,omitempty"`
		UpdateTime     string `json:"last_update,omitempty"`
		Coupon         string `json:"coupon.omitempty"`
	}
)
