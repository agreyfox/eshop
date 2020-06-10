- Capture notification

```
{
"id": "WH-58D329510W468432D-8HN650336L201105X",
"create_time": "2019-02-14T21:50:07.940Z",
"resource_type": "capture",
"event_type": "PAYMENT.CAPTURE.COMPLETED",
"summary": "Payment completed for \$ 2.51 USD",
"resource": {
"amount": {
"currency_code": "USD",
"value": "2.51"
},
"seller_protection": {
"status": "ELIGIBLE",
"dispute_categories": [
"ITEM_NOT_RECEIVED",
"UNAUTHORIZED_TRANSACTION"
]
},
"update_time": "2019-02-14T21:49:58Z",
"create_time": "2019-02-14T21:49:58Z",
"final_capture": true,
"seller_receivable_breakdown": {
"gross_amount": {
"currency_code": "USD",
"value": "2.51"
},
"paypal_fee": {
"currency_code": "USD",
"value": "0.37"
},
"net_amount": {
"currency_code": "USD",
"value": "2.14"
}
},
"links": [
{
"href": "https://api.paypal.com/v2/payments/captures/27M47624FP291604U",
"rel": "self",
"method": "GET"
},
{
"href": "https://api.paypal.com/v2/payments/captures/27M47624FP291604U/refund",
"rel": "refund",
"method": "POST"
},
{
"href": "https://api.paypal.com/v2/payments/authorizations/7W5147081L658180V",
"rel": "up",
"method": "GET"
}
],
"id": "27M47624FP291604U",
"status": "COMPLETED"
},
"links": [
{
"href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-58D329510W468432D-8HN650336L201105X",
"rel": "self",
"method": "GET",
"encType": "application/json"
},
{
"href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-58D329510W468432D-8HN650336L201105X/resend",
"rel": "resend",
"method": "POST",
"encType": "application/json"
}
],
"event_version": "1.0",
"resource_version": "2.0"
}
```

- dispute resolved notification

```

{
"id": "WH-6SA34813406280445-6M160383LS6155516",
"create_time": "2018-06-21T13:57:08.000Z",
"resource_type": "dispute",
"event_type": "CUSTOMER.DISPUTE.RESOLVED",
"summary": "A dispute was resolved with case # PP-000-042-663-135",
"resource": {
"reason": "MERCHANDISE_OR_SERVICE_NOT_RECEIVED",
"dispute_channel": "INTERNAL",
"create_time": "2018-06-21T13:35:44.000Z",
"dispute_id": "PP-000-042-663-135",
"dispute_life_cycle_stage": "CHARGEBACK",
"disputed_transactions": [
{
"seller_transaction_id": "00D10444LD479031K",
"seller": {
"merchant_id": "RD465XN5VS364",
"name": "Test Store"
},
"items": [],
"seller_protection_eligible": true
}
],
"update_time": "2018-06-21T13:55:58.000Z",
"seller_response_due_date": "2018-07-06T13:55:58.000Z",
"messages": [
{
"posted_by": "BUYER",
"time_posted": "2018-06-21T13:35:52.000Z",
"content": "qwqwqwq"
},
{
"posted_by": "SELLER",
"time_posted": "2018-06-21T13:41:36.000Z",
"content": "Escalating to paypal"
}
],
"links": [
{
"href": "https://api.paypal.com/v1/customer/disputes/PP-000-042-663-135",
"rel": "self",
"method": "GET"
},
{
"href": "https://api.paypal.com/v1/customer/disputes/PP-000-042-663-135/appeal",
"rel": "appeal",
"method": "POST"
}
],
"dispute_amount": {
"currency_code": "USD",
"value": "3.00"
},
"dispute_outcome": {
"outcome_code": "CANCELED_BY_BUYER"
},
"status": "RESOLVED"
},
"links": [
{
"href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-6SA34813406280445-6M160383LS6155516",
"rel": "self",
"method": "GET",
"encType": "application/json"
},
{
"href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-6SA34813406280445-6M160383LS6155516/resend",
"rel": "resend",
"method": "POST",
"encType": "application/json"
}
],
"event_version": "1.0"
}
```

- Refund completed notification

```
{
  "id": "WH-1GE84257G0350133W-6RW800890C634293G",
  "create_time": "2018-08-15T19:14:04.543Z",
  "resource_type": "refund",
  "event_type": "PAYMENT.CAPTURE.REFUNDED",
  "summary": "A $ 0.99 USD capture payment was refunded",
  "resource": {
    "seller_payable_breakdown": {
      "gross_amount": {
        "currency_code": "USD",
        "value": "0.99"
      },
      "paypal_fee": {
        "currency_code": "USD",
        "value": "0.02"
      },
      "net_amount": {
        "currency_code": "USD",
        "value": "0.97"
      },
      "total_refunded_amount": {
        "currency_code": "USD",
        "value": "1.98"
      }
    },
    "amount": {
      "currency_code": "USD",
      "value": "0.99"
    },
    "update_time": "2018-08-15T12:13:29-07:00",
    "create_time": "2018-08-15T12:13:29-07:00",
    "links": [
      {
        "href": "https://api.paypal.com/v2/payments/refunds/1Y107995YT783435V",
        "rel": "self",
        "method": "GET"
      },
      {
        "href": "https://api.paypal.com/v2/payments/captures/0JF852973C016714D",
        "rel": "up",
        "method": "GET"
      }
    ],
    "id": "1Y107995YT783435V",
    "status": "COMPLETED"
  },
  "links": [
    {
      "href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-1GE84257G0350133W-6RW800890C634293G",
      "rel": "self",
      "method": "GET",
      "encType": "application/json"
    },
    {
      "href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-1GE84257G0350133W-6RW800890C634293G/resend",
      "rel": "resend",
      "method": "POST",
      "encType": "application/json"
    }
  ],
  "event_version": "1.0",
  "resource_version": "2.0"
}
```
