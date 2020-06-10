PayPal develop

1 server side or client side.
2 payment or order
3 when to capture if need
3 using library
4 https://www.youtube.com/watch?v=AtZGoueL4Vs paypal with vue integration
5 https://www.youtube.com/watch?v=7k03jobKGXM for nodejs paypal example
6 use invoice by paypal or self.

a payment call:
{
"intent": "sale",
"redirect_urls": {
"return_url": "https://www.paypal.com/apex/api/redirect/success/expressCheckout/executeApprovedPayment?isInteractive=true&userId=1588986188030&productId=4",
"cancel_url": "https://www.paypal.com/apex/api/redirect/success/expressCheckout/createPayment?isInteractive=true&userId=1588986188030&productId=4"
},
"payer": {
"payment_method": "paypal"
},
"transactions": [
{
"amount": {
"total": "30.11",
"currency": "USD",
"details": {
"subtotal": "30.00",
"tax": "0.07",
"shipping": "0.03",
"handling_fee": "1.00",
"insurance": "0.01",
"shipping_discount": "-1.00"
}
},
"description": "The payment transaction description.",
"item_list": {
"items": [
{
"name": "hat",
"sku": "1",
"price": "3.00",
"currency": "USD",
"quantity": "5",
"description": "Brown hat.",
"tax": "0.01"
},
{
"name": "handbag",
"sku": "product34",
"price": "15.00",
"currency": "USD",
"quantity": "1",
"description": "Black handbag.",
"tax": "0.02"
}
]
}
}
]
}
the creation result is
{
"id": "1W6412894C358040B",
"intent": "CAPTURE",
"purchase_units": [
{
"reference_id": "default",
"amount": {
"currency_code": "USD",
"value": "30.11",
"breakdown": {
"item_total": {
"currency_code": "USD",
"value": "30.11"
}
}
},
"payee": {
"email_address": "sb-opduc1687278@business.example.com",
"merchant_id": "VRS5GBP9ETXXU"
},
"description": "note",
"invoice_id": "xzzzzzz",
"items": [
{
"name": "hat111",
"unit_amount": {
"currency_code": "USD",
"value": "3.00"
},
"quantity": "5",
"description": "Brown hat. for human",
"category": "DIGITAL_GOODS"
},
{
"name": "handbag222",
"unit_amount": {
"currency_code": "USD",
"value": "15.11"
},
"quantity": "1",
"description": "Black handbag. for spagati",
"category": "DIGITAL_GOODS"
}
],
"shipping": {
"name": {
"full_name": "Brian Robinson"
},
"address": {
"address_line_1": "4th Floor",
"admin_area_2": "San Jose",
"admin_area_1": "CA",
"postal_code": "300984",
"country_code": "US"
}
}
}
],
"payer": {
"email_address": "jihua.gao@mgmail.com"
},
"create_time": "2020-05-18T05:32:07Z",
"links": [
{
"href": "https://api.sandbox.paypal.com/v2/checkout/orders/1W6412894C358040B",
"rel": "self",
"method": "GET"
},
{
"href": "https://www.sandbox.paypal.com/checkoutnow?token=1W6412894C358040B",
"rel": "approve",
"method": "GET"
},
{
"href": "https://api.sandbox.paypal.com/v2/checkout/orders/1W6412894C358040B",
"rel": "update",
"method": "PATCH"
},
{
"href": "https://api.sandbox.paypal.com/v2/checkout/orders/1W6412894C358040B/capture",
"rel": "capture",
"method": "POST"
}
],
"status": "CREATED"
}


The Request 的说明

{
  "payer":"paypal",
  "email":"jihua.gao@mgmail.com",
  "item_list": [
        {
          "amount":{
                "value": "30.11",
                "currency_code": "USD",
                "breakdown":{
                  "item_total":{
                    "currency_code":"USD",
                    "value":"30.11"
                  }
                }
          },
          "invoice_id":"xzzzzzz",
          "description":"note",
          "shipping": {
            "name": {
              "full_name":"Brian Robinson"
            },
            "address":{
              "address_line_1": "4th Floor",
              "address_line_2":"",
              "admin_area_2": "San Jose",
              "admin_area_1": "CA",
              "country_code": "US",
              "postal_code":"300984"
            }
          },
          "items":[
          {
            "name": "hat111",
            "description": "Brown hat. for human",
            "unit_amount":{
              "currency_code":"USD",
              "value":"3"
            },
            "quantity": 5,
            "category": "DIGITAL_GOODS"
          },
          {
            "name": "handbag222",
            "description": "Black handbag. for spagati",
            "quantity": 1,
            "unit_amount":{
              "currency_code":"USD",
              "value":"15.11"
            },
            "category": "DIGITAL_GOODS"
          }
          ]
        }
  ] 
}