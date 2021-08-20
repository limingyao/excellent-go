package text

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type options struct {
	maxLen int
}

var (
	defaultOptions = options{
		maxLen: 10,
	}
)

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fo *funcOption) apply(o *options) {
	fo.f(o)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func WithMaxLength(maxLen int) Option {
	return newFuncOption(func(o *options) {
		o.maxLen = maxLen
	})
}

// Readable output human readable protobuf
func Readable(msg proto.Message, opts ...Option) string {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	msg.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		fmt.Println(descriptor.FullName(), descriptor.IsList(), descriptor.IsMap(), descriptor.IsPacked(), descriptor.IsPlaceholder())
		fmt.Println(descriptor.Kind())
		fmt.Println(value)
		return true
	})

	fmt.Println(protojson.Format(msg))

	return ""
}

func readable(msg proto.Message, opts ...Option) map[string]interface{} {
	buffer := make(map[string]interface{})
	msg.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {

		descriptor.IsList()
		descriptor.IsMap()

		switch descriptor.Kind() {
		case protoreflect.BoolKind:
		case protoreflect.EnumKind:
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
			protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
			protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		case protoreflect.FloatKind, protoreflect.DoubleKind:
		case protoreflect.StringKind:
		case protoreflect.BytesKind:
		case protoreflect.MessageKind:
		case protoreflect.GroupKind:
		}

		if descriptor.Name().IsValid() {
			buffer[string(descriptor.Name())] = value
		}
		return true
	})
	return buffer
}

// Compact output after replace large objects
func Compact() string {
	return ""
}
