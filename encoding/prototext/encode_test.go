package prototext_test

import (
	"testing"

	"github.com/limingyao/excellent-go/encoding/prototext"
)

func TestEncode(t *testing.T) {
	prototext.NewMarshaler()
	prototext.NewMarshaler(prototext.WithStringLimit(6))
}
