package transbank

import "webtyp.com/model"

func ExportMapStatus(status string) string {
	return mapStatus(status)
}

func ExportEncodeAmount(amount int64, currency string) (float64, error) {
	return encodeAmount(amount, currency)
}

func ExportDecodeAmount(amount float64, currency string) (int64, error) {
	return decodeAmount(amount, currency)
}

func (w *Webpay) ExportCall(method, path string, in model.Encodable, out model.Decodable) error {
	return w.call(method, path, in, out)
}

func OpenForTest(s string) {
    Open(s)
}
