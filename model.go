package transbank

import "webtyp.com/model"

// Modelos del JSON de la API REST v1.2 de Webpay Plus. Los nombres de campo
// son los de Transbank; model_orm.go lo genera ormc.

var CreateRequestModel = model.Definition{
	Name: "create_request",
	Fields: model.Fields{
		{Name: "buy_order", Type: model.Text()},
		{Name: "session_id", Type: model.Text()},
		{Name: "amount", Type: model.Float()},
		{Name: "return_url", Type: model.Text()},
	},
}

var CreateResponseModel = model.Definition{
	Name: "create_response",
	Fields: model.Fields{
		{Name: "token", Type: model.Text()},
		{Name: "url", Type: model.Text()},
	},
}

var CardDetailModel = model.Definition{
	Name: "card_detail",
	Fields: model.Fields{
		{Name: "card_number", Type: model.Text()},
	},
}

var TransactionResponseModel = model.Definition{
	Name: "transaction_response",
	Fields: model.Fields{
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

var RefundRequestModel = model.Definition{
	Name: "refund_request",
	Fields: model.Fields{
		{Name: "amount", Type: model.Float()},
	},
}

var RefundResponseModel = model.Definition{
	Name: "refund_response",
	Fields: model.Fields{
		{Name: "type", Type: model.Text()},
		{Name: "authorization_code", Type: model.Text()},
		{Name: "authorization_date", Type: model.Text()},
		{Name: "balance", Type: model.Float()},
		{Name: "nullified_amount", Type: model.Float()},
		{Name: "response_code", Type: model.Int()},
	},
}

var APIErrorModel = model.Definition{
	Name: "api_error",
	Fields: model.Fields{
		{Name: "error_message", Type: model.Text()},
	},
}
