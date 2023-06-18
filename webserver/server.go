package webserver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/limingyao/excellent-go/encoding/prototext"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptorOfDebug() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		if md, ok := metadata.FromIncomingContext(ctx); tracingId == "" && ok {
			sessionIds := md.Get(tracingCtxKey)
			if len(sessionIds) > 0 {
				tracingId = sessionIds[0]
			}
		}

		logger := log.WithField(tracingFieldNameKey, tracingId)
		if req != nil {
			reqText := prototext.CompactTextString(req.(proto.Message))
			logger.Infof("method: %s, request: %s", info.FullMethod, reqText)
		}

		start := time.Now()
		reply, err := handler(ctx, req)
		if err != nil {
			logger.Infof("method: %s, cost: %s, err: [%v]", info.FullMethod, time.Since(start), err)
		} else if reply != nil {
			rspText := prototext.CompactTextString(reply.(proto.Message))
			logger.Infof("method: %s, cost: %s, response: %s", info.FullMethod, time.Since(start), rspText)
		}

		return reply, err
	}
}

func UnaryServerInterceptorOfSessionId() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		if md, ok := metadata.FromIncomingContext(ctx); tracingId == "" && ok {
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

		ctx = context.WithValue(ctx, tracingCtxKey, tracingId)
		reply, err := handler(ctx, req)

		// inject tracing_id
		if value := reflect.ValueOf(reply); value.Kind() == reflect.Ptr {
			if elem := value.Elem(); elem.IsValid() {
				if field := elem.FieldByName(tracingFieldName); field.IsValid() {
					field.SetString(tracingId)
				}
			}
		}

		return reply, err
	}
}

func UnaryServerInterceptorOfRecovery() grpc.UnaryServerInterceptor {
	return grpcrecovery.UnaryServerInterceptor(
		grpcrecovery.WithRecoveryHandler(
			func(p interface{}) error {
				_, _ = os.Stderr.Write([]byte(fmt.Sprintf("panic recovered: %s", p)))
				debug.PrintStack()
				return status.Errorf(codes.Internal, "%s", p)
			},
		),
	)
}

func UnaryServerInterceptorOfContext() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("incoming metadata invalid")
		}
		// inject tracing_id
		if ctx.Value(tracingCtxKey) == nil {
			sessionIds := md.Get(tracingCtxKey)
			if len(sessionIds) > 0 {
				ctx = context.WithValue(ctx, tracingCtxKey, sessionIds[0])
			} else {
				if value := reflect.ValueOf(req); value.Kind() == reflect.Ptr {
					if elem := value.Elem(); elem.IsValid() {
						if field := elem.FieldByName(tracingFieldName); field.IsValid() {
							ctx = context.WithValue(ctx, tracingCtxKey, field.String())
						}
					}
				}
			}
		}
		// inject client ip
		if ctx.Value(clientIpCtxKey) == nil {
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
			// https://cloud.google.com/appengine/docs/flexible/python/reference/request-headers
			found := false
			if ips := md.Get("X-Forwarded-For"); len(ips) > 0 {
				found = true
				ctx = context.WithValue(ctx, clientIpCtxKey, strings.TrimSpace(strings.Split(ips[0], ",")[0]))
			}
			if ips := md.Get("X-Real-Ip"); !found && len(ips) > 0 {
				found = true
				ctx = context.WithValue(ctx, clientIpCtxKey, strings.TrimSpace(ips[0]))
			}
			if ips := md.Get("X-Appengine-Remote-Addr"); !found && len(ips) > 0 {
				found = true
				ctx = context.WithValue(ctx, clientIpCtxKey, strings.TrimSpace(ips[0]))
			}
		}
		return handler(ctx, req)
	}
}
