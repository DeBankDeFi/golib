package util

import (
	"context"

	ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc/metadata"
)

const (
	TraceID = "trace_id"
	AppName = "AppName"
	User    = "user"
)

// GetTraceIDFromContext lookup trace_id string from incoming context.
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx != nil && ctx.Value(TraceID) != nil {
		traceId, ok := ctx.Value(TraceID).(string)
		if ok {
			return traceId
		}
	}
	return ""
}

// GetTraceIDPairsFromContext lookup trace_id string from incoming context and
// return constant TraceID and realistic trace_id string.
// The returned trace_id pairs commonly used for `zap.String(lu.GetTraceIDPairsFromContext(ctx))`
func GetTraceIDPairsFromContext(ctx context.Context) (string, string) {
	return TraceID, GetTraceIDFromContext(ctx)
}

// GetTraceIDFromGRPCContext lookup trace_id string from incoming context that produced by grpc.
func GetTraceIDFromGRPCContext(ctx context.Context) string {
	return getMetadataFromGRPCContext(ctx, TraceID)
}

// GetAppNameFromGRPCContext lookup app_name string from incoming context that produced by grpc calling.
func GetAppNameFromGRPCContext(ctx context.Context) string {
	return getMetadataFromGRPCContext(ctx, AppName)
}

// GetAppNameFromGRPCContext lookup username string from incoming context that produced by grpc calling.
func GetUserFromGRPCContext(ctx context.Context) string {
	var user string
	if ctxtags.Extract(ctx).Has(User) {
		val := ctxtags.Extract(ctx).Values()
		user = val[User].(string)
	}
	return user
}

func getMetadataFromGRPCContext(ctx context.Context, typ string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		tid := md.Get(typ)
		if len(tid) == 1 {
			return tid[0]
		}
	}
	if val, ok := ctx.Value(typ).(string); ok {
		return val
	}
	return ""
}

// GetTraceIDPairsFromGRPCContext lookup trace_id string from incoming context that produced by grpc.
// return constant TraceID and realistic trace_id string.
// The returned trace_id pairs commonly used for `zap.String(lu.GetTraceIDPairsFromContext(ctx))`
func GetTraceIDPairsFromGRPCContext(ctx context.Context) (string, string) {
	return TraceID, GetTraceIDFromGRPCContext(ctx)
}

// SetTraceIDToContext attach a trace_id string to an existing context.
func SetTraceIDToContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceID, traceID)
}

// OverwriteAppNameToGrpcContext overwrite application name to a incoming context which produced by grpc.
func OverwriteAppNameToGrpcContext(ctx context.Context, appName string) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		tid := md.Get(AppName)
		if len(tid) == 1 {
			md.Set(AppName, appName)
		}
	}
	return ctx
}
