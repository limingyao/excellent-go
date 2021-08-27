package proto

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type ProtoMarshaller struct {
	runtime.ProtoMarshaller
}

// ContentType always returns "application/octet-stream".
func (*ProtoMarshaller) ContentType(interface{}) string {
	return binding.MIMEPROTOBUF
}
