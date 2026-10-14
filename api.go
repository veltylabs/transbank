package transbank

import (
	"webtyp.com/fmt"
	"webtyp.com/json"
	fmt2 "fmt"
	"webtyp.com/model"
	"webtyp.com/fetch"
)

func (w *Webpay) call(method, path string, in model.Encodable, out model.Decodable) error {
	url := w.cfg.BaseURL + path

	var req *fetch.Request
	switch method {
	case "GET":
		req = fetch.Get(url)
	case "POST":
		req = fetch.Post(url)
	case "PUT":
		req = fetch.Put(url)
	case "DELETE":
		req = fetch.Delete(url)
	default:
		return fmt.Err("transbank: unsupported method")
	}

	if in != nil && !in.IsNil() {
		var reqBody []byte
		err := json.Encode(in, &reqBody)
		if err != nil {
			return err
		}

		req.Body(reqBody)
		req.Header("Content-Type", "application/json")
	}

	req.Header("Tbk-Api-Key-Id", w.cfg.CommerceCode)
	req.Header("Tbk-Api-Key-Secret", w.cfg.APIKey)
	req.Timeout(w.cfg.TimeoutMs)

	errCh := make(chan error, 1)
	req.Send(func(res *fetch.Response, err error) {
		if err != nil {
			errCh <- fmt.Err("transbank: network: ", err.Error())
			return
		}

		if res.Status < 200 || res.Status >= 300 {
			var apiErr APIError
			bodyBytes := res.Body()
			if err := json.Decode(bodyBytes, &apiErr); err == nil && apiErr.ErrorMessage != "" {
				errCh <- fmt2.Errorf("transbank: HTTP %d: %s", res.Status, apiErr.ErrorMessage)
				return
			}
			errCh <- fmt2.Errorf("transbank: HTTP %d", res.Status)
			return
		}

		if out != nil {
			if err := json.Decode(res.Body(), out); err != nil {
				errCh <- fmt.Err("transbank: bad response: ", err.Error())
				return
			}
		}
		errCh <- nil
	})

	return <-errCh
}
