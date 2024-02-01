package interceptors

import (
	"context"
	"fmt"
	"os"
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
	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoRequest interface {
	GetSessionId() string
}

const (
	protoSessionIdFieldName = "SessionId"

	metadataForwardedKey  = "X-Forwarded-For"
	metadataRealIpKey     = "X-Real-Ip"
	metadataRemoteAddrKey = "X-Appengine-Remote-Addr"
	metadataSessionIdKey  = "X-Session-Id"

	CtxClientKey    = "client_ip"  // context keys
	CtxSessionIdKey = "session_id" // context keys
)

// UnaryServerInterceptorOfRecovery recovery
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

// UnaryServerInterceptorOfContext inject session_id, client_ip to context
func UnaryServerInterceptorOfContext() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "%s", "incoming metadata invalid")
		}

		// inject session id
		if ctx.Value(CtxSessionIdKey) == nil {
			sessionIds := md.Get(metadataSessionIdKey)
			if len(sessionIds) > 0 {
				ctx = context.WithValue(ctx, CtxSessionIdKey, sessionIds[0])
			} else if req, ok := req.(protoRequest); ok && req != nil {
				ctx = context.WithValue(ctx, CtxSessionIdKey, req.GetSessionId())
			}
		}

		// inject client ip
		if ctx.Value(CtxClientKey) == nil {
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
			// https://cloud.google.com/appengine/docs/flexible/python/reference/request-headers
			found := false
			if ips := md.Get(metadataForwardedKey); len(ips) > 0 {
				found = true
				ctx = context.WithValue(ctx, CtxClientKey, strings.TrimSpace(strings.Split(ips[0], ",")[0]))
			}
			if ips := md.Get(metadataRealIpKey); !found && len(ips) > 0 {
				found = true
				ctx = context.WithValue(ctx, CtxClientKey, strings.TrimSpace(ips[0]))
			}
			if ips := md.Get(metadataRemoteAddrKey); !found && len(ips) > 0 {
				found = true
				ctx = context.WithValue(ctx, CtxClientKey, strings.TrimSpace(ips[0]))
			}
		}

		return handler(ctx, req)
	}
}

// UnaryServerInterceptorOfSessionId inject session_id to response
func UnaryServerInterceptorOfSessionId() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		sessionId, ok := ctx.Value(CtxSessionIdKey).(string)
		if !ok || len(sessionId) < 1 {
			sessionId = uuid.New().String()
			ctx = context.WithValue(ctx, CtxSessionIdKey, sessionId)
		}

		resp, err = handler(ctx, req)

		// inject session id
		p := proto.MessageReflect(resp.(proto.Message))
		if field := p.Descriptor().Fields().ByName(protoSessionIdFieldName); field != nil {
			p.Set(field, protoreflect.ValueOfString(sessionId))
		}

		return resp, err
	}
}

// UnaryServerInterceptorOfDebug print request, response
func UnaryServerInterceptorOfDebug() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		sessionId, ok := ctx.Value(CtxSessionIdKey).(string)
		if !ok || len(sessionId) < 1 {
			sessionId = uuid.New().String()
			ctx = context.WithValue(ctx, CtxSessionIdKey, sessionId)
		}
		logger := log.WithField(CtxSessionIdKey, sessionId)

		// print request
		if req != nil {
			text := prototext.CompactTextString(req.(proto.Message))
			logger.WithField("method", info.FullMethod).Infof("request: %s", text)
		}

		start := time.Now()
		resp, err = handler(ctx, req)
		cost := time.Since(start)

		// print response
		if err != nil {
			logger.WithError(err).WithField("method", info.FullMethod).Infof("cost: %v", cost)
		} else if resp != nil {
			text := prototext.CompactTextString(resp.(proto.Message))
			logger.WithField("method", info.FullMethod).Infof("response: %s, cost: %v", text, cost)
		}

		return resp, err
	}
}
