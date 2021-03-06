package prototext_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/limingyao/excellent-go/encoding/prototext"
	"github.com/limingyao/excellent-go/test"
)

func TestReadable(t *testing.T) {
	hello := &test.Hello{}
	hello.SessionId = uuid.New().String()
	hello.InstanceId = 10000
	hello.Names = append(hello.Names, "hello", "中文")
	hello.Version = "v1.0"
	hello.Data = &test.Data{}
	hello.Data.Video = []byte("video")
	hello.Data.Images = make(map[int32][]byte)
	hello.Data.Images[0] = []byte("first image")
	hello.Data.Images[1] = []byte("second image")
	hello.Data.Images[2] = []byte("third image")

	t.Log(proto.MarshalTextString(hello))
	t.Log(prototext.MarshalTextString(hello))
	t.Log(proto.CompactTextString(hello))
	t.Log(prototext.CompactTextString(hello))

	m := prototext.NewMarshaler(prototext.WithCompact())
	t.Log(m.Text(hello))

	m = prototext.NewMarshaler(prototext.WithCompact(), prototext.WithStringLimit(6))
	t.Log(m.Text(hello))
}

func TestCompact(t *testing.T) {

}
