package tests

import (
	"testing"
	"webtyp.com/json"
	"github.com/veltylabs/transbank"
	"webtyp.com/fmt"
)

func TestModels_RoundTrip(t *testing.T) {
	req := transbank.CreateRequest{
		BuyOrder:  "O-1",
		SessionId: "S-1",
		Amount:    1000,
		ReturnUrl: "http://return",
	}
	_ = req.Schema()
	_ = req.Pointers()
	_ = req.Validate(0)
	req.IsNil()

	res := transbank.CreateResponse{
		Token: "T-1",
		Url:   "http://url",
	}
	_ = res.Schema()
	_ = res.Pointers()
	_ = res.Validate(0)
	res.IsNil()

	com := transbank.TransactionResponse{
		Vci:                "VCI-1",
		Amount:             1000,
		Status:             "AUTHORIZED",
		BuyOrder:           "O-1",
		SessionId:          "S-1",
		CardDetail:         transbank.CardDetail{CardNumber: "1234"},
		AccountingDate:     "1234",
		TransactionDate:    "1234",
		AuthorizationCode:  "1234",
		PaymentTypeCode:    "1234",
		ResponseCode:       0,
		InstallmentsAmount: 0,
		InstallmentsNumber: 0,
		Balance:            0,
	}
	_ = com.Schema()
	_ = com.Pointers()
	_ = com.Validate(0)
	com.IsNil()

	card := transbank.CardDetail{
		CardNumber: "1234",
	}
	_ = card.Schema()
	_ = card.Pointers()
	_ = card.Validate(0)
	card.IsNil()

	refReq := transbank.RefundRequest{
		Amount: 1000,
	}
	_ = refReq.Schema()
	_ = refReq.Pointers()
	_ = refReq.Validate(0)
	refReq.IsNil()

	refRes := transbank.RefundResponse{
		Type:              "REVERSED",
		AuthorizationCode: "1234",
		AuthorizationDate: "1234",
		NullifiedAmount:   1000,
		Balance:           0,
		ResponseCode:      0,
	}
	_ = refRes.Schema()
	_ = refRes.Pointers()
	_ = refRes.Validate(0)
	refRes.IsNil()

	apiErr := transbank.APIError{
		ErrorMessage: "Error",
	}
	_ = apiErr.Schema()
	_ = apiErr.Pointers()
	_ = apiErr.Validate(0)
	apiErr.IsNil()
}

func TestModels_RoundTrip2(t *testing.T) {
	req := transbank.CreateRequest{
		BuyOrder:  "O-1",
		SessionId: "S-1",
		Amount:    1000,
		ReturnUrl: "http://return",
	}
	buf := fmt.GetConv()
	_ = json.Encode(&req, buf)
	var decReq transbank.CreateRequest
	_ = json.Decode(buf.Bytes(), &decReq)
	buf.PutConv()

	res := transbank.CreateResponse{
		Token: "T-1",
		Url:   "http://url",
	}
	buf = fmt.GetConv()
	_ = json.Encode(&res, buf)
	var decRes transbank.CreateResponse
	_ = json.Decode(buf.Bytes(), &decRes)
	buf.PutConv()

	com := transbank.TransactionResponse{
		Vci:                "VCI-1",
		Amount:             1000,
		Status:             "AUTHORIZED",
		BuyOrder:           "O-1",
		SessionId:          "S-1",
		CardDetail:         transbank.CardDetail{CardNumber: "1234"},
		AccountingDate:     "1234",
		TransactionDate:    "1234",
		AuthorizationCode:  "1234",
		PaymentTypeCode:    "1234",
		ResponseCode:       0,
		InstallmentsAmount: 0,
		InstallmentsNumber: 0,
		Balance:            0,
	}
	buf = fmt.GetConv()
	_ = json.Encode(&com, buf)
	var decCom transbank.TransactionResponse
	_ = json.Decode(buf.Bytes(), &decCom)
	buf.PutConv()

	refRes := transbank.RefundResponse{
		Type:              "REVERSED",
		AuthorizationCode: "1234",
		AuthorizationDate: "1234",
		NullifiedAmount:   1000,
		Balance:           0,
		ResponseCode:      0,
	}
	buf = fmt.GetConv()
	_ = json.Encode(&refRes, buf)
	var decRefRes transbank.RefundResponse
	_ = json.Decode(buf.Bytes(), &decRefRes)
	buf.PutConv()

	apiErr := transbank.APIError{
		ErrorMessage: "Error",
	}
	buf = fmt.GetConv()
	_ = json.Encode(&apiErr, buf)
	var decApiErr transbank.APIError
	_ = json.Decode(buf.Bytes(), &decApiErr)
	buf.PutConv()

	refReq := transbank.RefundRequest{
		Amount: 1000,
	}
	buf = fmt.GetConv()
	_ = json.Encode(&refReq, buf)
	var decRefReq transbank.RefundRequest
	_ = json.Decode(buf.Bytes(), &decRefReq)
	buf.PutConv()
}

func TestModels_Lists_At(t *testing.T) {
	reqList := transbank.CreateRequestList{{BuyOrder: "1"}}
	_ = reqList.At(0)

	resList := transbank.CreateResponseList{{Token: "1"}}
	_ = resList.At(0)

	comList := transbank.TransactionResponseList{{Vci: "1"}}
	_ = comList.At(0)

	cardList := transbank.CardDetailList{{CardNumber: "1"}}
	_ = cardList.At(0)

	refReqList := transbank.RefundRequestList{{Amount: 1}}
	_ = refReqList.At(0)

	refResList := transbank.RefundResponseList{{Type: "1"}}
	_ = refResList.At(0)

	apiErrList := transbank.APIErrorList{{ErrorMessage: "1"}}
	_ = apiErrList.At(0)
}

func TestModels_Lists_EncodeDecode2(t *testing.T) {
	reqList := transbank.CreateRequestList{{BuyOrder: "O-1"}, {BuyOrder: "O-2"}}
	buf := fmt.GetConv()
	json.Encode(&reqList, buf)
	var newReqList transbank.CreateRequestList
	json.Decode(buf.Bytes(), &newReqList)
	buf.PutConv()

	resList := transbank.CreateResponseList{{Token: "T-1"}, {Token: "T-2"}}
	buf = fmt.GetConv()
	json.Encode(&resList, buf)
	var newResList transbank.CreateResponseList
	json.Decode(buf.Bytes(), &newResList)
	buf.PutConv()

	comList := transbank.TransactionResponseList{{Vci: "1"}, {Vci: "2"}}
	buf = fmt.GetConv()
	json.Encode(&comList, buf)
	var newComList transbank.TransactionResponseList
	json.Decode(buf.Bytes(), &newComList)
	buf.PutConv()

	cardList := transbank.CardDetailList{{CardNumber: "1"}, {CardNumber: "2"}}
	buf = fmt.GetConv()
	json.Encode(&cardList, buf)
	var newCardList transbank.CardDetailList
	json.Decode(buf.Bytes(), &newCardList)
	buf.PutConv()

	refReqList := transbank.RefundRequestList{{Amount: 1}, {Amount: 2}}
	buf = fmt.GetConv()
	json.Encode(&refReqList, buf)
	var newRefReqList transbank.RefundRequestList
	json.Decode(buf.Bytes(), &newRefReqList)
	buf.PutConv()

	refResList := transbank.RefundResponseList{{Type: "1"}, {Type: "2"}}
	buf = fmt.GetConv()
	json.Encode(&refResList, buf)
	var newRefResList transbank.RefundResponseList
	json.Decode(buf.Bytes(), &newRefResList)
	buf.PutConv()

	apiErrList := transbank.APIErrorList{{ErrorMessage: "1"}, {ErrorMessage: "2"}}
	buf = fmt.GetConv()
	json.Encode(&apiErrList, buf)
	var newApiErrList transbank.APIErrorList
	json.Decode(buf.Bytes(), &newApiErrList)
	buf.PutConv()
}

func TestModels_Lists_Append(t *testing.T) {
	reqList := transbank.CreateRequestList{}
	_ = reqList.Append()

	resList := transbank.CreateResponseList{}
	_ = resList.Append()

	comList := transbank.TransactionResponseList{}
	_ = comList.Append()

	cardList := transbank.CardDetailList{}
	_ = cardList.Append()

	refReqList := transbank.RefundRequestList{}
	_ = refReqList.Append()

	refResList := transbank.RefundResponseList{}
	_ = refResList.Append()

	apiErrList := transbank.APIErrorList{}
	_ = apiErrList.Append()
}

func TestModels_Lists_EncodeDecode3(t *testing.T) {
	reqList := transbank.CreateRequestList{{BuyOrder: "O-1"}, {BuyOrder: "O-2"}}
	reqList.EncodeFields(nil)
	reqList.DecodeFields(nil)

	resList := transbank.CreateResponseList{{Token: "T-1"}, {Token: "T-2"}}
	resList.EncodeFields(nil)
	resList.DecodeFields(nil)

	comList := transbank.TransactionResponseList{{Vci: "1"}, {Vci: "2"}}
	comList.EncodeFields(nil)
	comList.DecodeFields(nil)

	cardList := transbank.CardDetailList{{CardNumber: "1"}, {CardNumber: "2"}}
	cardList.EncodeFields(nil)
	cardList.DecodeFields(nil)

	refReqList := transbank.RefundRequestList{{Amount: 1}, {Amount: 2}}
	refReqList.EncodeFields(nil)
	refReqList.DecodeFields(nil)

	refResList := transbank.RefundResponseList{{Type: "1"}, {Type: "2"}}
	refResList.EncodeFields(nil)
	refResList.DecodeFields(nil)

	apiErrList := transbank.APIErrorList{{ErrorMessage: "1"}, {ErrorMessage: "2"}}
	apiErrList.EncodeFields(nil)
	apiErrList.DecodeFields(nil)
}
