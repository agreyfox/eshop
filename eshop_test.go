package main

import (
	"fmt"
	"testing"

	"github.com/agreyfox/eshop/payment"
	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/payment/paypal"
	"github.com/agreyfox/eshop/system/db"
	"github.com/go-zoo/bone"
)

func TestSend(t *testing.T) {
	//mail := admin.NewMailClient("grimmnanettehjbb@gmail.com", "qweasdzxC123^&*")
	//	mail.Send("标题1", "邮箱内容1", "jihua.gao@gmail.com") //邮件标题 邮件内容 需要发送到的邮箱地址
}

func TestSendConfirmEmail(t *testing.T) {
	db.Init()
	data.SendConfirmEmail("xde32455", "Notice", "23.9", "USD", "jihua.gao@gmail.com")
}

func TestGetLogContent(t *testing.T) {
	db.Init()
	mainMux := bone.New()
	payment.InitialPayment(db.Store(), mainMux)
	data, err := data.GetLogContent("55065193b261eb7da0c16c92d3dbbd3a", paypal.PaypalCreated)
	fmt.Println(data, err)
}

func TestGetPurchaseContent(t *testing.T) {
	db.Init()
	mainMux := bone.New()
	payment.InitialPayment(db.Store(), mainMux)

	data := paypal.GetPurchaseContent("55065193b261eb7da0c16c92d3dbbd3a")
	//data.GetLogContent("55065193b261eb7da0c16c92d3dbbd3a", paypal.PaypalCreated)
	fmt.Println(data)
}

func TestSaveTransationDetail(t *testing.T) {
	db.Init()
	mainMux := bone.New()
	payment.InitialPayment(db.Store(), mainMux)
	payment.Run("paypal")
	paypal.SaveTransationDetail("2WA67358PY5856315")

}
