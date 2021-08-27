package proto

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// ProtoMarshaller is a Marshaller which marshals/unmarshals into/from serialize proto bytes
type ProtoMarshaller struct {
	runtime.ProtoMarshaller
}

// ContentType always returns "application/octet-stream".
func (*ProtoMarshaller) ContentType(interface{}) string {
	return binding.MIMEPROTOBUF
}
