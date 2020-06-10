package skrill

import (
	"strings"
)

const ()

// Config is configration to initiate Skrill client
type Config struct {
	URL        string
	Email      string
	MerchantID string
	SecretWord string
}

// PrepareParam describes describes the payment source used to make Prepare
type PrepareParam struct {
	// Merchant details
	PayToEmail           string   `json:"pay_to_email"`
	RecipientDescription string   `json:"recipient_description,omitempty"`
	TransactionID        string   `json:"transaction_id,omitempty"`
	ReturnURL            string   `json:"return_url,omitempty"`
	ReturnURLText        string   `json:"return_url_text,omitempty"`
	ReturnURLTarget      Target   `json:"return_url_target,omitempty"`
	CancelURL            string   `json:"cancel_url,omitempty"`
	CancelURLTarget      string   `json:"cancel_url_target,omitempty"`
	StatusURL            string   `json:"status_url,omitempty"`
	StatusURL2           string   `json:"status_url2,omitempty"`
	Language             Language `json:"language,omitempty"`
	LogoURL              string   `json:"logo_url,omitempty"`
	PrepareOnly          string   `json:"prepare_only"`
	SID                  string   `json:"sid,omitempty"`
	RID                  string   `json:"rid,omitempty"`
	ExtRefID             string   `json:"ext_ref_id,omitempty"`
	DynamicDescriptor    string   `json:"dynamic_descriptor,omitempty"`
	MerchantFields       string   `json:"merchant_fields,omitempty`
	OrderID              string   `json:"order_id"`
	PayFromEmail         string   `json:"pay_from_email,omitempty"`
	FirstName            string   `json:"firstname,omitempty"`
	LastName             string   `json:"lastname,omitempty"`
	DateOfBirth          string   `json:"date_of_birth,omitempty"`
	Address              string   `json:"address,omitempty"`
	Address2             string   `json:"address2,omitempty"`
	PhoneNumber          string   `json:"phone_number,omitempty"`
	PostalCode           string   `json:"postal_code,omitempty"`
	City                 string   `json:"city,omitempty"`
	Country              string   `json:"country,omitempty"`
	NetellerAccount      string   `json:"neteller_account,omitempty"`
	NetellerSecureID     string   `json:"neteller_secure_id,omitempty"`
	PaymentMethods       string   `json:"payment_methods,omitempty"`
	// Payment details
	Amount             float64  `json:"amount"`
	Currency           Currency `json:"currency"`
	Amount2Description string   `json:"amount2_description,omitempty"`
	Amount2            float64  `json:"amount2,omitempty"`
	Amount3Description string   `json:"amount3_description,omitempty"`
	Amount3            float64  `json:"amount3,omitempty"`
	Amount4Description string   `json:"amount4_description,omitempty"`
	Amount4            float64  `json:"amount4,omitempty"`
	Detail1Description string   `json:"detail1_description,omitempty"`
	Detail1Text        string   `json:"detail1_text,omitempty"`
	Detail2Description string   `json:"detail2_description,omitempty"`
	Detail2Text        string   `json:"detail2_text,omitempty"`
	Detail3Description string   `json:"detail3_description,omitempty"`
	Detail3Text        string   `json:"detail3_text,omitempty"`
	Detail4Description string   `json:"detail4_description,omitempty"`
	Detail4Text        string   `json:"detail4_text,omitempty"`
	Detail5Description string   `json:"detail5_description,omitempty"`
	Detail5Text        string   `json:"detail5_text,omitempty"`
}

// StatusResponse describes a request body from Skrill API when changing status of payment
type StatusResponse struct {
	PayToEmail       string   `json:"pay_to_email"`
	PayFromEmail     string   `json:"pay_from_email"`
	MerchantID       string   `json:"merchant_id"`
	CustomerID       string   `json:"customer_id,omitempty"`
	TransactionID    string   `json:"transaction_id,omitempty"`
	MbTransactionID  string   `json:"mb_transaction_id"`
	MbAmount         float64  `json:"mb_amount,string"`
	MbCurrency       Currency `json:"mb_currency"`
	Status           Status   `json:"status,string"`
	FailedReasonCode Code     `json:"failed_reason_code,string,omitempty"`
	Md5Sig           string   `json:"md5sig"`
	Sha2Sig          string   `json:"sha2sig"`
	Amount           float64  `json:"amount,string"`
	Currency         Currency `json:"currency"`
	NetellerID       string   `json:"neteller_id,omitempty"`
	PaymentType      string   `json:"payment_type,omitempty"`
	OrderID          string   `json:"order_id,omitempty"`
}

// Status represents Status of StatusResponse
type Status int

// Status list
const (
	SkrillCreated    Status = 1
	SkrillProcessed  Status = 2
	SkrillPending    Status = 0
	SkrillCancelled  Status = -1
	SkrillFailed     Status = -2
	SkrillChargeback Status = -3
)

func (status Status) String() string {
	switch status {
	case SkrillCreated:
		return "created"
	case SkrillProcessed:
		return "processed"
	case SkrillPending:
		return "pending"
	case SkrillCancelled:
		return "cancelled"
	case SkrillFailed:
		return "failed"
	case SkrillChargeback:
		return "chargeback"
	default:
		return "unkown"
	}
}

// Code describes failed_reason_code from Skrill API
type Code int

// Code list
const (
	CardIssue                            Code = 1  // Referred by Card Issuer
	InvalidMerchant                      Code = 2  // Invalid Merchant
	PickupCard                           Code = 3  // Pick‐up card
	Declined                             Code = 4  // Declined by Card Issuer
	InsufficientFUnds                    Code = 5  // Insufficient funds
	TransactionFailed                    Code = 6  // Transaction failed
	IncorrectPIN                         Code = 7  // Incorrect PIN
	PINTriesExceed                       Code = 8  // PIN tries exceed ‐ card blocked
	InvalidTransaction                   Code = 9  // Invalid Transaction
	TransactioLimitExceeded              Code = 10 // Transaction frequency limit exceeded
	InvalidAmount                        Code = 11 // Invalid Amount/ Amount too high /Limit Exceeded
	InvalidCreditCardOrBank              Code = 12 // Invalid credit card or bank account
	InvalidCard                          Code = 13 // Invalid card Issuer
	DuplicateTransaction                 Code = 15 // Duplicate transaction
	RetryTransaction                     Code = 19 // Retry transaction
	CardExpired                          Code = 24 // Card expired
	RequestedFuncNotAvailable            Code = 27 // Requested function not available
	LostCard                             Code = 28 // Lost/stolen card
	FormatFailure                        Code = 30 // Format Failure
	WrongSecurityCode                    Code = 32 // Card Security Code (CVV2/CVC2) Check Failed
	IllegalTransaction                   Code = 34 // Illegal Transaction
	CardRestricted                       Code = 37 // Card restricted by Card Issuer
	SecrityViolation                     Code = 38 // Security violation
	CardBlocked                          Code = 42 // Card blocked by Card Issuer
	BankOrNetworkUnavailable             Code = 44 // Card Issuing Bank or Network is not available
	ProcessingError                      Code = 45 // Processing error ‐ card type is not processed by the authorization centre
	SytemError                           Code = 51 // System error
	TransactionNotPermittedByAcquirer    Code = 58 // Transaction not permitted by acquirer
	TransactionNotPermittedForCardholder Code = 63 // Transaction not permitted for cardholder
	Wrong3DSVerification                 Code = 70 // Customer failed 3DS verification
	FraudRulesDeclined                   Code = 80 // Fraud rules declined
	ErrorWithProvider                    Code = 98 // Error in communication with provider
	Other                                Code = 99 // Other
)

// Target is target value for return_url_target and cancel_url_target
type Target int

// Target lists
const (
	Top    Target = 1
	Parent Target = 2
	Self   Target = 3
	Blank  Target = 4
)

// Language describes language which can be used to make a payment through Skrill API
type Language string

// Language List
const (
	BG Language = "BG"
	CS Language = "CS"
	DA Language = "DA"
	DE Language = "DE"
	EL Language = "EL"
	EN Language = "EN"
	ES Language = "ES"
	FI Language = "FI"
	FR Language = "FR"
	IT Language = "IT"
	ZH Language = "ZH"
	NL Language = "NL"
	PL Language = "PL"
	RO Language = "RO"
	RU Language = "RU"
	SV Language = "SV"
	TR Language = "TR"
	JA Language = "JA"
)

// Currency is the list of supported currencies.
type Currency string

// Currencies which can be used for a payment through Skrill
const (
	EUR Currency = "EUR" // Euro TWD Taiwan Dollar
	USD Currency = "USD" // U.S. Dollar THB Thailand Baht
	GBP Currency = "GBP" // British Pound CZK Czech Koruna
	HKD Currency = "HKD" // Hong Kong Dollar HUF Hungarian Forint
	SGD Currency = "SGD" // Singapore Dollar BGN Bulgarian Leva
	JPY Currency = "JPY" // Japanese Yen PLN Polish Zloty
	CAD Currency = "CAD" // Canadian Dollar ISK Iceland Krona
	AUD Currency = "AUD" // Australian Dollar INR Indian Rupee
	CHF Currency = "CHF" // Swiss Franc KRW South‐Korean Won
	DKK Currency = "DKK" // Danish Krone ZAR South‐African Rand
	SEK Currency = "SEK" // Swedish Krona RON Romanian Leu New
	NOK Currency = "NOK" // Norwegian Krone HRK Croatian Kuna
	ILS Currency = "ILS" // Israeli Shekel JOD Jordanian Dinar
	MYR Currency = "MYR" // Malaysian Ringgit OMR Omani Rial
	NZD Currency = "NZD" // New Zealand Dollar RSD Serbian Dinar
	TRY Currency = "TRY" // New Turkish Lira TND Tunisian Dinar
	AED Currency = "AED" // Utd. Arab Emir. Dirham BHD Bahraini Dinar
	MAD Currency = "MAD" // Moroccan Dirham KWD Kuwaiti Dinar
	QAR Currency = "QAR" // Qatari Rial
	SAR Currency = "SAR" // Saudi Riy
)

// GetCurrencyCode Return Currency struct by string
func GetCurrencyCode(str string) Currency {
	cc := strings.ToUpper(str)
	switch cc {
	case "EUR": // Euro TWD Taiwan Dollar
		return EUR
	case "USD": // U.S. Dollar THB Thailand Baht
		return USD
	case "GBP": // British Pound CZK Czech Koruna
		return GBP
	case "HKD": // Hong Kong Dollar HUF Hungarian Forint
		return EUR
	case "SGD": // Singapore Dollar BGN Bulgarian Leva
		return EUR
	case "JPY": // Japanese Yen PLN Polish Zloty
		return JPY
	case "CAD": // Canadian Dollar ISK Iceland Krona
		return CAD
	case "AUD": // Australian Dollar INR Indian Rupee
		return AUD
	case "CHF": // Swiss Franc KRW South‐Korean Won
		return CHF
	case "DKK": //= Danish Krone ZAR South‐African Rand
		return DKK
	case "SEK": // Swedish Krona RON Romanian Leu New
		return SEK
	case "NOK": // Norwegian Krone HRK Croatian Kuna
		return NOK
	case "ILS": // Israeli Shekel JOD Jordanian Dinar
		return ILS
	case "MYR": // Malaysian Ringgit OMR Omani Rial
		return MYR
	case "NZD": // New Zealand Dollar RSD Serbian Dinar
		return NZD
	case "TRY": // New Turkish Lira TND Tunisian Dinar
		return TRY
	case "AED": // Utd. Arab Emir. Dirham BHD Bahraini Dinar
		return AED
	case "MAD": // Moroccan Dirham KWD Kuwaiti Dinar
		return MAD
	case "QAR": // Qatari Rial
		return QAR
	case "SAR": // Saudi Riy
		return SAR
	default:
		return USD
	}
}
