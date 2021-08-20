package text

import (
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"testing"
)

func TestReadable(t *testing.T) {
	hello := &Hello{}
	hello.SessionId = uuid.New().String()
	hello.InstanceId = 10000
	hello.Names = append(hello.Names, "hello", "world")
	hello.Version = "v1.0"
	hello.Data = &Data{}
	hello.Data.Video = []byte("video")
	hello.Data.Images = make(map[int32][]byte)
	hello.Data.Images[0] = []byte("first image")
	hello.Data.Images[1] = []byte("second image")
	hello.Data.Images[2] = []byte("third image")

	t.Log(proto.CompactTextString(hello))
	t.Log(CompactTextString(hello))
}

func TestCompact(t *testing.T) {

}
