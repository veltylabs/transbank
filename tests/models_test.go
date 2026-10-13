package tests

import (
	"bytes"
	"testing"
	"webtyp.com/json"
	"github.com/veltylabs/transbank"
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

	reqList := transbank.CreateRequestList{}
	_ = reqList.IsNil()
	_ = reqList.Len()
	_ = reqList.Append()

	res := transbank.CreateResponse{
		Token: "T-1",
		Url:   "http://url",
	}
	_ = res.Schema()
	_ = res.Pointers()
	_ = res.Validate(0)
	res.IsNil()

	resList := transbank.CreateResponseList{}
	_ = resList.IsNil()
	_ = resList.Len()
	_ = resList.Append()

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

	comList := transbank.TransactionResponseList{}
	_ = comList.IsNil()
	_ = comList.Len()
	_ = comList.Append()

	card := transbank.CardDetail{
		CardNumber: "1234",
	}
	_ = card.Schema()
	_ = card.Pointers()
	_ = card.Validate(0)
	card.IsNil()

	cardList := transbank.CardDetailList{}
	_ = cardList.IsNil()
	_ = cardList.Len()
	_ = cardList.Append()

	refReq := transbank.RefundRequest{
		Amount: 1000,
	}
	_ = refReq.Schema()
	_ = refReq.Pointers()
	_ = refReq.Validate(0)
	refReq.IsNil()

	refReqList := transbank.RefundRequestList{}
	_ = refReqList.IsNil()
	_ = refReqList.Len()
	_ = refReqList.Append()

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

	refResList := transbank.RefundResponseList{}
	_ = refResList.IsNil()
	_ = refResList.Len()
	_ = refResList.Append()

	apiErr := transbank.APIError{
		ErrorMessage: "Error",
	}
	_ = apiErr.Schema()
	_ = apiErr.Pointers()
	_ = apiErr.Validate(0)
	apiErr.IsNil()

	apiErrList := transbank.APIErrorList{}
	_ = apiErrList.IsNil()
	_ = apiErrList.Len()
	_ = apiErrList.Append()

}

func TestModels_RoundTrip2(t *testing.T) {
	req := transbank.CreateRequest{
		BuyOrder:  "O-1",
		SessionId: "S-1",
		Amount:    1000,
		ReturnUrl: "http://return",
	}
	var buf bytes.Buffer
	_ = json.Encode(&req, &buf)
	var decReq transbank.CreateRequest
	_ = json.Decode(buf.Bytes(), &decReq)

	res := transbank.CreateResponse{
		Token: "T-1",
		Url:   "http://url",
	}
	buf.Reset()
	_ = json.Encode(&res, &buf)
	var decRes transbank.CreateResponse
	_ = json.Decode(buf.Bytes(), &decRes)

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
	buf.Reset()
	_ = json.Encode(&com, &buf)
	var decCom transbank.TransactionResponse
	_ = json.Decode(buf.Bytes(), &decCom)

	refRes := transbank.RefundResponse{
		Type:              "REVERSED",
		AuthorizationCode: "1234",
		AuthorizationDate: "1234",
		NullifiedAmount:   1000,
		Balance:           0,
		ResponseCode:      0,
	}
	buf.Reset()
	_ = json.Encode(&refRes, &buf)
	var decRefRes transbank.RefundResponse
	_ = json.Decode(buf.Bytes(), &decRefRes)

	apiErr := transbank.APIError{
		ErrorMessage: "Error",
	}
	buf.Reset()
	_ = json.Encode(&apiErr, &buf)
	var decApiErr transbank.APIError
	_ = json.Decode(buf.Bytes(), &decApiErr)

	refReq := transbank.RefundRequest{
		Amount: 1000,
	}
	buf.Reset()
	_ = json.Encode(&refReq, &buf)
	var decRefReq transbank.RefundRequest
	_ = json.Decode(buf.Bytes(), &decRefReq)
}

func TestModels_Lists_EncodeDecode2(t *testing.T) {
	reqList := transbank.CreateRequestList{{BuyOrder: "O-1"}, {BuyOrder: "O-2"}}
	var outReq bytes.Buffer
	json.Encode(&reqList, &outReq)
	var newReqList transbank.CreateRequestList
	json.Decode(outReq.Bytes(), &newReqList)

	resList := transbank.CreateResponseList{{Token: "T-1"}, {Token: "T-2"}}
	var outRes bytes.Buffer
	json.Encode(&resList, &outRes)
	var newResList transbank.CreateResponseList
	json.Decode(outRes.Bytes(), &newResList)

	comList := transbank.TransactionResponseList{{Vci: "1"}, {Vci: "2"}}
	var outCom bytes.Buffer
	json.Encode(&comList, &outCom)
	var newComList transbank.TransactionResponseList
	json.Decode(outCom.Bytes(), &newComList)

	cardList := transbank.CardDetailList{{CardNumber: "1"}, {CardNumber: "2"}}
	var outCard bytes.Buffer
	json.Encode(&cardList, &outCard)
	var newCardList transbank.CardDetailList
	json.Decode(outCard.Bytes(), &newCardList)

	refReqList := transbank.RefundRequestList{{Amount: 1}, {Amount: 2}}
	var outRefReq bytes.Buffer
	json.Encode(&refReqList, &outRefReq)
	var newRefReqList transbank.RefundRequestList
	json.Decode(outRefReq.Bytes(), &newRefReqList)

	refResList := transbank.RefundResponseList{{Type: "1"}, {Type: "2"}}
	var outRefRes bytes.Buffer
	json.Encode(&refResList, &outRefRes)
	var newRefResList transbank.RefundResponseList
	json.Decode(outRefRes.Bytes(), &newRefResList)

	apiErrList := transbank.APIErrorList{{ErrorMessage: "1"}, {ErrorMessage: "2"}}
	var outApiErr bytes.Buffer
	json.Encode(&apiErrList, &outApiErr)
	var newApiErrList transbank.APIErrorList
	json.Decode(outApiErr.Bytes(), &newApiErrList)
}
