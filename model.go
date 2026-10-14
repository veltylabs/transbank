package transbank

import "webtyp.com/model"

type CreateRequest struct {
	BuyOrder  string  `json:"buy_order"`
	SessionId string  `json:"session_id"`
	Amount    float64 `json:"amount"`
	ReturnUrl string  `json:"return_url"`
}

var CreateRequestModel = model.Definition{
	Name: "CreateRequest",
	Fields: []model.Field{
		{Name: "buy_order", Type: model.Text()},
		{Name: "session_id", Type: model.Text()},
		{Name: "amount", Type: model.Float()},
		{Name: "return_url", Type: model.Text()},
	},
}

type CreateRequestList []CreateRequest

type CreateResponse struct {
	Token string `json:"token"`
	Url   string `json:"url"`
}

var CreateResponseModel = model.Definition{
	Name: "CreateResponse",
	Fields: []model.Field{
		{Name: "token", Type: model.Text()},
		{Name: "url", Type: model.Text()},
	},
}

type CreateResponseList []CreateResponse

type TransactionResponse struct {
	Vci                string     `json:"vci"`
	Amount             float64    `json:"amount"`
	Status             string     `json:"status"`
	BuyOrder           string     `json:"buy_order"`
	SessionId          string     `json:"session_id"`
	CardDetail         CardDetail `json:"card_detail"`
	AccountingDate     string     `json:"accounting_date"`
	TransactionDate    string     `json:"transaction_date"`
	AuthorizationCode  string     `json:"authorization_code"`
	PaymentTypeCode    string     `json:"payment_type_code"`
	ResponseCode       int64      `json:"response_code"`
	InstallmentsAmount float64    `json:"installments_amount"`
	InstallmentsNumber int64      `json:"installments_number"`
	Balance            float64    `json:"balance"`
}

var TransactionResponseModel = model.Definition{
	Name: "TransactionResponse",
	Fields: []model.Field{
		{Name: "vci", Type: model.Text()},
		{Name: "amount", Type: model.Float()},
		{Name: "status", Type: model.Text()},
		{Name: "buy_order", Type: model.Text()},
		{Name: "session_id", Type: model.Text()},
		{Name: "card_detail", Type: model.Struct(&CardDetailModel)},
		{Name: "accounting_date", Type: model.Text()},
		{Name: "transaction_date", Type: model.Text()},
		{Name: "authorization_code", Type: model.Text()},
		{Name: "payment_type_code", Type: model.Text()},
		{Name: "response_code", Type: model.Int()},
		{Name: "installments_amount", Type: model.Float()},
		{Name: "installments_number", Type: model.Int()},
		{Name: "balance", Type: model.Float()},
	},
}

type TransactionResponseList []TransactionResponse

type CardDetail struct {
	CardNumber string `json:"card_number"`
}

var CardDetailModel = model.Definition{
	Name: "CardDetail",
	Fields: []model.Field{
		{Name: "card_number", Type: model.Text()},
	},
}

type CardDetailList []CardDetail

type RefundRequest struct {
	Amount float64 `json:"amount"`
}

var RefundRequestModel = model.Definition{
	Name: "RefundRequest",
	Fields: []model.Field{
		{Name: "amount", Type: model.Float()},
	},
}

type RefundRequestList []RefundRequest

type RefundResponse struct {
	Type              string  `json:"type"`
	AuthorizationCode string  `json:"authorization_code"`
	AuthorizationDate string  `json:"authorization_date"`
	NullifiedAmount   float64 `json:"nullified_amount"`
	Balance           float64 `json:"balance"`
	ResponseCode      int64   `json:"response_code"`
}

var RefundResponseModel = model.Definition{
	Name: "RefundResponse",
	Fields: []model.Field{
		{Name: "type", Type: model.Text()},
		{Name: "authorization_code", Type: model.Text()},
		{Name: "authorization_date", Type: model.Text()},
		{Name: "nullified_amount", Type: model.Float()},
		{Name: "balance", Type: model.Float()},
		{Name: "response_code", Type: model.Int()},
	},
}

type RefundResponseList []RefundResponse

type APIError struct {
	ErrorMessage string `json:"error_message"`
}

var APIErrorModel = model.Definition{
	Name: "APIError",
	Fields: []model.Field{
		{Name: "error_message", Type: model.Text()},
	},
}

type APIErrorList []APIError
