package chain

import (
	context "context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger // General logger package chain

func InitLogger(baseLogger *zap.SugaredLogger) {
	// Init package's logger here with distinct name here
	logger = baseLogger.Named("chain")
}

type contextIDType int

const (
	requestIDKey contextIDType = iota
)

// WithRequestID adds a random requestID to a context
func WithRequestID(ctx context.Context, iden Identifier) context.Context {
	id := iden.GetUUID()
	if len(id) == 0 {
		randUUID, _ := uuid.NewRandom()
		id = randUUID.String()
	}
	return context.WithValue(ctx, requestIDKey, id)
}

// Logger returns a logger attached with a requestID if it's
// available in the context
func Logger(ctx context.Context) *zap.SugaredLogger {
	l := logger
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		l = l.With("requestID", requestID)
	}
	return l
}

type Identifier interface {
	GetUUID() string
}
