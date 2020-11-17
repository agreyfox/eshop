package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/agreyfox/eshop/system/admin"
	"github.com/agreyfox/eshop/system/email"
	"github.com/boltdb/bolt"
)

/*
数据结构如下
records.db 中 payments 是 root bucket 相当于table
			"orderid"是子bucket
				“state":key
				orderentry ： value
*/

// SaveRequest to save the data in payment
func SaveOrderRequest(r *UserSubmitOrderRequest) error {

	//log.RequestTime = time.Now().Unix()
	//state := log.PaymentState

	//ID := r.OrderID // TODO: need find better id

	tx, err := PaymentDBHandler.Begin(true)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer tx.Rollback()

	root := tx.Bucket([]byte(DBRequest))

	if buf, err := json.Marshal(r); err != nil {
		logger.Error(err)
		return PaymentErrInputData
	} else if err := root.Put([]byte(r.OrderID), buf); err != nil {
		logger.Error(err)
		return PaymentErrSave
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return PaymentErrSave
	}

	logger.Info("Save user order request payment data done. ")

	return nil
}

func GetRequestByID(orderid string) (*UserSubmitOrderRequest, error) {

	ret := make([]byte, 0)
	err := PaymentDBHandler.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(DBRequest))

		ret = b.Get([]byte(orderid))

		return nil
	})
	if err != nil {
		logger.Error("Get order state error ", err)
		return nil, err
	}
	rr := UserSubmitOrderRequest{}

	datasrc := bytes.NewReader(ret)
	err = json.NewDecoder(datasrc).Decode(&rr)
	return &rr, err
}

///=========================================================================
/*
数据结构如下
records.db 中 payments 是 root bucket 相当于table
			"orderid"是子bucket
				“state":key
				orderentry ： value
*/

// SaveRequest to save the data in payment
func SavePaymentLog(log *PaymentLog) bool {

	log.RequestTime = time.Now().Unix()
	state := log.PaymentState

	ID := log.OrderID // TODO: need find better id

	tx, err := PaymentLogHandler.Begin(true)
	if err != nil {
		logger.Error(err)
		return false
	}
	defer tx.Rollback()

	root := tx.Bucket([]byte(DBLogName))
	// Setup the order bucket.
	bkt, err := root.CreateBucketIfNotExists([]byte(ID))
	if err != nil {
		logger.Error(err)
		return false
	}

	if buf, err := json.Marshal(log); err != nil {
		logger.Error(err)
		return false
	} else if err := bkt.Put([]byte(state), buf); err != nil {
		logger.Error(err)
		return false
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}

	//logger.Info("Save payment data done. ")

	return true
}

// return a orderid 's key
func GetLog(orderid string) ([]string, error) {
	///ID := orderid + IDMaker + vendor
	ret := make([]string, 0)
	err := PaymentLogHandler.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(DBLogName)).Bucket([]byte(orderid))

		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			ret = append(ret, string(k))
		}

		return nil
	})
	if err != nil {
		logger.Error("Get order state error ", err)
		return ret, err
	}
	return ret, nil
}

// return a id's status content
func GetLogContent(id string, status string) ([]byte, error) {
	val := &bytes.Buffer{}
	err := PaymentLogHandler.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(DBLogName)).Bucket([]byte(id))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if strings.EqualFold(string(k[:]), status) {
				val.Write(v)
				break
			}
		}
		return nil
	})
	if err != nil {
		logger.Error("Get payment log content  error ", err)
		return val.Bytes(), err
	}
	//logger.Debug(string(val.Bytes()))
	return val.Bytes(), nil
}

func GetRequestByState(orderid string, state string) (*PaymentLog, error) {
	///ID := orderid + IDMaker + vendor
	ret := &PaymentLog{}
	logger.Debugf("Get Request log with data orderid %s and state %s", orderid, state)
	err := PaymentLogHandler.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(DBLogName)).Bucket([]byte(orderid))
		if b == nil {
			return errors.New("No such bucket/data!")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(k) == state {
				err := json.Unmarshal(v, ret)
				return err
			}
		}
		return errors.New("No such Data!")
	})
	if err != nil {
		logger.Error("Get order state error ", err)
		return ret, err
	}
	logger.Debug("done:", ret.BuyerEmail)
	return ret, nil
}

// SaveComplete to save the data in payment complete order
func CreateOrderComplete(order *Order) bool {

	order.RequestTime = fmt.Sprintf(time.Now().Format(time.RFC1123))
	ID := order.OrderID

	err := PaymentLogHandler.Update(func(tx *bolt.Tx) error {
		encoded, err := json.Marshal(order)
		if err != nil {
			return err
		}

		err = tx.Bucket([]byte(DBLogName)).Bucket([]byte(Complete)).Put([]byte(ID), encoded)

		return err
	})
	if err != nil {
		logger.Error("Save order data error ", err)
		return false
	}
	logger.Debug("Now to save the order into system db ...")
	return true
}

func GetOrder(vendor string, orderid string) (*Order, error) {
	ID := orderid
	ret := &Order{}
	err := PaymentLogHandler.View(func(tx *bolt.Tx) error {

		data := tx.Bucket([]byte(DBLogName)).Bucket([]byte(Complete)).Get([]byte(ID))

		if data != nil {
			_ = json.Unmarshal(data, ret)
		}
		return errors.New("Order Data is not exists!")
	})
	if err != nil {
		logger.Error("Get Order  data error ", err)
		return ret, err
	}
	return ret, nil
}

func GetRefundByEmail(email string) []PaymentLog {
	return []PaymentLog{}
}

func GetChargeBackByEmail(email string) []PaymentLog {
	return []PaymentLog{}
}

/* state should be
OrderCreated    = "已创建"
	OrderPaid       = "已付款"
	OrderInValidate = "待检验"
	OrderInDelivery = "待交付"
	OrderCompleted  = "已完成"
	OrderCancel     = "用户取消"
	OrderRefunded   = "用户退款"
	OrderDisputed   = "用户欺诈"
*/
// id is order id, so you need get real content order id by orderid first
func UpdateOrderByID(id, key, value string) bool {
	//updatetime := fmt.Sprintf(time.Now().Format(time.RFC1123))

	oid := admin.FindContentID("Order", id, "order_id")
	if oid == "" {
		logger.Debugf("record id not found with order id is %s", id)
		return false
	}
	i, err := admin.UpdateContent("Order", oid, key, []byte(value))
	if err != nil {
		logger.Error("Update order error: ", err)
		return false
	} else {
		logger.Debug("Update Order done with id:", i)
		return true
	}

	/* tx, err := SystemDBHandler.Begin(true)
	if err != nil {
		logger.Error(err)
		return false
	}
	defer tx.Rollback()

	orderBucket := tx.Bucket([]byte(OrderName))

	data := orderBucket.Get([]byte(oid))
	order := &Order{}
	err = json.Unmarshal(data, order)
	if err != nil {
		logger.Error(err)
		return false
	}
	order.UpdateTime = updatetime
	order.Status = state
	err = orderBucket.Delete([]byte(id))

	if buf, err := json.Marshal(order); err != nil {
		logger.Error(err)
		return false
	} else if err := orderBucket.Put([]byte(id), buf); err != nil {
		logger.Error(err)
		return false
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}
	*/
	// /logger.Info("Update Order status done. ")

}

//check wether the give state is valid
func IsValidStatus(state string) bool {
	if state == OrderCreated || state == OrderPaid || state == OrderInValidate || state == OrderInDelivery || state == OrderCompleted || state == OrderCancel || state == OrderRefunded || state == OrderDisputed {

		return true
	}
	logger.Error("Status is not valide !")
	return false
}

// SendConfimEmail is formal email sending
func SendConfirmEmail(orderid, content, price, currency string, mailaddr string) {
	tomail := []string{"18901882538@189.cn"}

	tomail = append(tomail, removeQuote(mailaddr))
	htmlbody := `
	<p>
	Dear,<br>
		This is content of your order below:<br>
	========================================<br>
	Order Number : %s<br>
	Order Contain : %s <br>
	Price: %s %s <br>
	========================================<br>
	Thanks for your consistent support our delivery guy is coming to trade. ifyou don't get any whisper in 15 mins, Please contract out livechat.<br>
	This is an automated email form system mainbox, Please do not reply. <br>
	
	
	======= <br>

	wwww.egpal.com<br>
	</p>
	`

	logger.Debugf("Try to send email to %s\n", removeQuote(mailaddr))
	fmt.Println(tomail)
	msg := email.Email{
		//From: admin.MailUser,
		To:       tomail,
		Subject:  "EGPal EShop Order notification letter",
		TextBody: fmt.Sprintf("You have make purchase order(%s),Total %s%sWe will check and delivery soon\n Thanks!", orderid, price, currency),
		//HtmlBody: fmt.Sprintf("<h1>EGPal Purchase Confirmation</h1> <p>You have make order(%s) Successfully. We will check and delivery soon</p>", orderid),
		HtmlBody: fmt.Sprintf(htmlbody, orderid, content, price, currency),
	}
	res, err := email.Send(&msg)
	if err != nil {
		fmt.Printf("An Error Occurred: %s\n", err)
	}

	if res.Data.Succeeded == 1 {
		logger.Debugf("Order Sent Successfully: %s\n", orderid)
	} else {
		logger.Errorf("Order Sent with error: %v\n", res.Data)
	}
}

// remove string head and tail quote
func removeQuote(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
