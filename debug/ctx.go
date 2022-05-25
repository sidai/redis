package debug

import (
	"context"
	"fmt"
	"time"
)

type CtxKey string

const (
	CtxKeyID        CtxKey = "ID"
	CtxKeyStartTime CtxKey = "StartTime"
)

type DebugCtx struct {
	context.Context
}

func NewDebugCtx(ctx context.Context, id string) context.Context {
	ctx = context.WithValue(ctx, CtxKeyID, id)
	ctx = context.WithValue(ctx, CtxKeyStartTime, time.Now())

	return &DebugCtx{
		Context: ctx,
	}
}

func GetCtxID(ctx context.Context) string {
	if id, ok := ctx.Value(CtxKeyID).(string); ok {
		return fmt.Sprintf("%-15v", id)
	}

	return "Background     "
}

func GetCtxElapsedTime(ctx context.Context) string {
	if startTime, ok := ctx.Value(CtxKeyStartTime).(time.Time); ok {
		return fmt.Sprintf("%5dms", time.Now().Sub(startTime).Milliseconds())
	}

	return "No Time"
}
