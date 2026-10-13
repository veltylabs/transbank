package tests

import (
	"testing"
	"github.com/veltylabs/transbank"
	"webtyp.com/model"
)

type badDecodable struct{}
func (b *badDecodable) DecodeFields(r model.FieldReader) {}
func (b *badDecodable) IsNil() bool { return false }

func TestCall_JsonDecodeNil(t *testing.T) {
	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      "http://localhost",
		TimeoutMs:    1000,
	})

	err := wp.ExportCall("POST", "/test", nil, &badDecodable{})
	if err == nil {
	}
}

type badEncodable struct{}
func (b *badEncodable) EncodeFields(w model.FieldWriter) {}
func (b *badEncodable) IsNil() bool { return false }

func TestCall_JsonEncodeErr(t *testing.T) {
	wp := transbank.New(transbank.Config{
		BaseURL: "http://localhost",
	})
	_ = wp.ExportCall("POST", "/test", &badEncodable{}, nil)
}
