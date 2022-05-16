// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package proto provides functionality for handling protocol buffer messages.
// In particular, it provides marshaling and unmarshaling between a protobuf
// message and the binary wire format.
//
// See https://developers.google.com/protocol-buffers/docs/gotutorial for
// more information.
//
// Deprecated: Use the "google.golang.org/protobuf/proto" package instead.

package prototext

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// Message is a protocol buffer message.
//
// This is the v1 version of the message interface and is marginally better
// than an empty interface as it lacks any method to programatically interact
// with the contents of the message.
//
// A v2 message is declared in "google.golang.org/protobuf/proto".Message and
// exposes protobuf reflection as a first-class feature of the interface.
//
// To convert a v1 message to a v2 message, use the MessageV2 function.
// To convert a v2 message to a v1 message, use the MessageV1 function.
type Message = protoiface.MessageV1

// MessageReflect returns a reflective view for a message.
// It returns nil if m is nil.
func MessageReflect(m Message) protoreflect.Message {
	return protoimpl.X.MessageOf(m)
}

func isMessageSet(md protoreflect.MessageDescriptor) bool {
	ms, ok := md.(interface{ IsMessageSet() bool })
	return ok && ms.IsMessageSet()
}
