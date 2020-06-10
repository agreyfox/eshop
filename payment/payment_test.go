package payment

import (
	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/payment/skrill"
	"github.com/agreyfox/eshop/system/db"
	"github.com/go-zoo/bone"
	"testing"
)

func TestUpdateOrderByID(t *testing.T) {
	db.Init()

	id := "7W5147081L658180V"
	state := "待检验"
	data.UpdateOrderByID(id, "status", state)

}

func TestSaveNotify(t *testing.T) {
	db.Init()
	mainMux := bone.New()
	InitialPayment(db.Store(), mainMux)
	skrill.Start(mainMux)
	data := "order_id=2323mm3k4n&transaction_id=3195856960&mb_amount=1.3&amount=1.3&md5sig=B23743880D2FAE5D02F0205ABBF9B6FA&merchant_id=138853317&payment_type=WLT&mb_transaction_id=3195856960&mb_currency=USD&pay_from_email=18901882538%40189.cn&pay_to_email=e_raeb%40163.com&currency=USD&customer_id=139601073&status=2"
	skrill.SaveNotify([]byte(data))

}
