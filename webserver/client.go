package webserver

import (
	"context"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/limingyao/excellent-go/encoding/prototext"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	tracingFieldName    = "SessionId"
	tracingFieldNameKey = "session_id"

	tracingCtxKey  = "X-Session-Id"
	clientIpCtxKey = "X-Forwarded-For"
)

// SetTracing fieldName for proto field name, fieldNameKey for log key, ctxKey for context key
func SetTracing(fieldName, fieldNameKey, ctxKey string) {
	tracingFieldName = fieldName
	tracingFieldNameKey = fieldNameKey
	tracingCtxKey = ctxKey
}

func GetTracingCtxKey() string {
	return tracingCtxKey
}

func GetClientIpCtxKey() string {
	return clientIpCtxKey
}

func UnaryClientInterceptorOfSessionId() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// get tracing_id
		tracingId := ""
		found := true
		if value := reflect.ValueOf(req); value.Kind() == reflect.Ptr {
			if elem := value.Elem(); elem.IsValid() {
				if field := elem.FieldByName(tracingFieldName); field.IsValid() {
					tracingId = field.String()
				}
			}
		}
		if v, ok := ctx.Value(tracingCtxKey).(string); tracingId == "" && ok {
			found = false
			tracingId = v
		}
		if md, ok := metadata.FromOutgoingContext(ctx); tracingId == "" && ok {
			found = false
			if sessionIds := md.Get(tracingCtxKey); len(sessionIds) > 0 {
				tracingId = sessionIds[0]
			}
		}
		if tracingId == "" {
			found = false
			tracingId = uuid.New().String()
		}

		// inject tracing_id
		if value := reflect.ValueOf(req); !found && value.Kind() == reflect.Ptr {
			if elem := value.Elem(); elem.IsValid() {
				if field := elem.FieldByName(tracingFieldName); field.IsValid() {
					field.SetString(tracingId)
				}
			}
		}

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		md.Set(tracingCtxKey, tracingId)
		ctx = metadata.NewOutgoingContext(ctx, md)
		err := invoker(ctx, method, req, reply, cc, opts...)

		// inject tracing_id
		if value := reflect.ValueOf(reply); value.Kind() == reflect.Ptr {
			if elem := value.Elem(); elem.IsValid() {
				if field := elem.FieldByName(tracingFieldName); field.IsValid() {
					field.SetString(tracingId)
				}
			}
		}

		return err
	}
}

func UnaryClientInterceptorOfDebug() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// get tracing_id
		tracingId := ""
		if value := reflect.ValueOf(req); value.Kind() == reflect.Ptr {
			if elem := value.Elem(); elem.IsValid() {
				if field := elem.FieldByName(tracingFieldName); field.IsValid() {
					tracingId = field.String()
				}
			}
		}
		if v, ok := ctx.Value(tracingCtxKey).(string); tracingId == "" && ok {
			tracingId = v
		}
		if md, ok := metadata.FromOutgoingContext(ctx); tracingId == "" && ok {
			sessionIds := md.Get(tracingCtxKey)
			if len(sessionIds) > 0 {
				tracingId = sessionIds[0]
			}
		}

		logger := log.WithField(tracingFieldNameKey, tracingId)
		if req != nil {
			reqText := prototext.CompactTextString(req.(proto.Message))
			logger.Infof("client invoke method: %s, request: %s", method, reqText)
		}

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			logger.Infof("client invoke method: %s, cost: %s, err: [%v]", method, time.Since(start), err)
		} else if reply != nil {
			rspText := prototext.CompactTextString(reply.(proto.Message))
			logger.Infof("client invoke method: %s, cost: %s, response: %s", method, time.Since(start), rspText)
		}

		return err
	}
}

func UnaryClientInterceptorOfContext() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		tracingId := ""
		if v, ok := ctx.Value(tracingCtxKey).(string); ok {
			tracingId = v
		}
		md := map[string]string{
			tracingCtxKey: tracingId,
		}
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func ClientDialOptions(options ...grpc.DialOption) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(ClientMaxReceiveMessageSize),
			grpc.MaxCallSendMsgSize(ClientMaxSendMessageSize),
		),
		grpc.WithChainUnaryInterceptor(
			UnaryClientInterceptorOfSessionId(),
			UnaryClientInterceptorOfContext(),
			UnaryClientInterceptorOfDebug(),
		),
	}
	return append(opts, options...)
}
