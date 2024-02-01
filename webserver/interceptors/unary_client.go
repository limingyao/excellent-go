package interceptors

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/limingyao/excellent-go/encoding/prototext"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// UnaryClientInterceptorOfMetadata inject session_id to metadata
func UnaryClientInterceptorOfMetadata() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		sessionId, ok := ctx.Value(CtxSessionIdKey).(string)
		if !ok || len(sessionId) < 1 {
			sessionId = uuid.New().String()
			ctx = context.WithValue(ctx, CtxSessionIdKey, sessionId)
		}

		md := map[string]string{
			metadataSessionIdKey: sessionId,
		}
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))

		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

// UnaryClientInterceptorOfSessionId inject session_id to request
func UnaryClientInterceptorOfSessionId() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		sessionId, ok := ctx.Value(CtxSessionIdKey).(string)
		if !ok || len(sessionId) < 1 {
			sessionId = uuid.New().String()
			ctx = context.WithValue(ctx, CtxSessionIdKey, sessionId)
		}

		// inject session id
		p := proto.MessageReflect(req.(proto.Message))
		if field := p.Descriptor().Fields().ByName(protoSessionIdFieldName); field != nil {
			p.Set(field, protoreflect.ValueOfString(sessionId))
		}

		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

// UnaryClientInterceptorOfDebug print request, response
func UnaryClientInterceptorOfDebug() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		sessionId, ok := ctx.Value(CtxSessionIdKey).(string)
		if !ok || len(sessionId) < 1 {
			sessionId = uuid.New().String()
			ctx = context.WithValue(ctx, CtxSessionIdKey, sessionId)
		}
		logger := log.WithField(CtxSessionIdKey, sessionId)

		// print request
		if req != nil {
			text := prototext.CompactTextString(req.(proto.Message))
			logger.WithField("method", method).Infof("request: %s", text)
		}

		start := time.Now()
		err := invoker(ctx, method, req, resp, cc, opts...)
		cost := time.Since(start)

		// print response
		if err != nil {
			logger.WithError(err).WithField("method", method).Infof("cost: %v", cost)
		} else if resp != nil {
			text := prototext.CompactTextString(resp.(proto.Message))
			logger.WithField("method", method).Infof("response: %s, cost: %v", text, cost)
		}

		return err
	}
}
