package payment

import (
	"fmt"
	"testing"
	"time"

	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/payment/paypal"
	"github.com/agreyfox/eshop/payment/skrill"
	"github.com/agreyfox/eshop/system/db"
	"github.com/go-zoo/bone"
)

func TestUpdateOrderByID(t *testing.T) {
	db.Init("system.db")

	id := "7W5147081L658180V"
	state := "待检验"
	data.UpdateOrderByID(id, "status", state)

}

func TestGetnerateOrderID(t *testing.T) {
	db.Init("system.db")

	fmt.Println(data.GetShortOrderID())

}
func TestSaveNotify(t *testing.T) {
	db.Init("system.db")
	mainMux := bone.New()
	InitialPayment(db.Store(), mainMux)
	skrill.Start(mainMux)
	//	data := "order_id=2323mm3k4n&transaction_id=3195856960&mb_amount=1.3&amount=1.3&md5sig=B23743880D2FAE5D02F0205ABBF9B6FA&merchant_id=138853317&payment_type=WLT&mb_transaction_id=3195856960&mb_currency=USD&pay_from_email=18901882538%40189.cn&pay_to_email=e_raeb%40163.com&currency=USD&customer_id=139601073&status=2"
	//	skrill.SaveNotify([]byte(data))

}

func TestGetLogContent(t *testing.T) {
	db.Init("system.db")
	mainMux := bone.New()
	InitialPayment(db.Store(), mainMux)
	data, err := data.GetLogContent("55065193b261eb7da0c16c92d3dbbd3a", paypal.PaypalCreated)
	fmt.Println(data, err)
}

func TestGetTransationDetail(t *testing.T) {
	db.Init("system.db")
	mainMux := bone.New()
	InitialPayment(db.Store(), mainMux)
	//data, err := data.GetLogContent("55065193b261eb7da0c16c92d3dbbd3a", paypal.PaypalCreated)
	paypal.Start(mainMux)
	paypal.GetTransationDetail("31")
}

func TestUpdateOrderStatusByID(t *testing.T) {
	db.Init("system.db")
	mainMux := bone.New()
	InitialPayment(db.Store(), mainMux)
	//data, err := data.GetLogContent("55065193b261eb7da0c16c92d3dbbd3a", paypal.PaypalCreated)
	paypal.Start(mainMux)
	paypal.UpdateOrderStatusByID("31", "mmm")
}
func TestSaveOrderRequest(t *testing.T) {
	db.Init("system.db")
	mainMux := bone.New()
	InitialPayment(db.Store(), mainMux)
	//data, err := data.GetLogContent("55065193b261eb7da0c16c92d3dbbd3a", paypal.PaypalCreated)
	orderdata := data.UserSubmitOrderRequest{
		OrderID:     "1234567",
		OrderDate:   time.Now().Unix(),
		Email:       "jihua.gao@gmail.com",
		RequestInfo: "abc",
		IPAddr:      "127.0.0.1",
		Amount:      1.9,
		ItemList: []data.Item{
			{Game: "abs", Server: "server1", Category: "agreyfox", Product: "刀", UnitPrice: 12.3, Quantity: 2},
			{Game: "abc", Server: "server2", Category: "agreyfox", Product: "剑", UnitPrice: 1.3, Quantity: 1},
		},
	}

	fmt.Println(data.SaveOrderRequest(&orderdata))
	m, err := data.GetRequestByID("1234567")
	fmt.Println(m, err)

}
