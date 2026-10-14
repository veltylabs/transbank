package transbank

import (
	"webtyp.com/fetch"
	"webtyp.com/fmt"
	"webtyp.com/json"
	"webtyp.com/model"
)

// Ruta y cabeceras de la API REST v1.2 de Webpay Plus.
const (
	TransactionsPath  = "/rswebpaytransaction/api/webpay/v1.2/transactions"
	HeaderKeyID       = "Tbk-Api-Key-Id"
	HeaderKeySecret   = "Tbk-Api-Key-Secret"
	methodGet         = "GET"
	methodPost        = "POST"
	methodPut         = "PUT"
	refundsPathSuffix = "/refunds"
)

// call hace un pedido a la API y espera la respuesta (los métodos del Gateway
// son síncronos y solo corren en el servidor). in == nil: sin cuerpo.
func (w *Webpay) call(method, path string, in model.Encodable, out model.Decodable) error {
	url := w.cfg.BaseURL + path
	var req *fetch.Request
	switch method {
	case methodPost:
		req = fetch.Post(url)
	case methodPut:
		req = fetch.Put(url)
	default:
		req = fetch.Get(url)
	}
	req.ContentTypeJSON().
		Header(HeaderKeyID, w.cfg.CommerceCode).
		Header(HeaderKeySecret, w.cfg.APIKey).
		Timeout(w.cfg.TimeoutMs)
	if in != nil {
		var body []byte
		_ = json.Encode(in, &body) // los modelos generados no fallan al codificar
		req.Body(body)
	}

	done := make(chan error, 1)
	req.Send(func(res *fetch.Response, err error) {
		if err != nil { // C1
			done <- fmt.Err("transbank: network:", err.Error())
			return
		}
		if res.Status < 200 || res.Status > 299 { // C2
			var apiErr APIError
			if json.Decode(res.Body(), &apiErr) == nil && apiErr.ErrorMessage != "" { // C3
				done <- fmt.Errf("transbank: HTTP %d: %s", res.Status, apiErr.ErrorMessage)
				return
			}
			done <- fmt.Errf("transbank: HTTP %d", res.Status)
			return
		}
		if json.Decode(res.Body(), out) != nil { // C4
			done <- ErrBadResponse
			return
		}
		done <- nil
	})
	return <-done
}
