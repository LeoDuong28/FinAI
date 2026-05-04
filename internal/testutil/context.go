package testutil

import (
	"context"

	"github.com/google/uuid"

	"github.com/nghiaduong/finai/internal/middleware"
)

func ContextWithUserID(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, userID)
}

func ContextWithJTI(ctx context.Context, jti string) context.Context {
	return context.WithValue(ctx, middleware.JTIKey, jti)
}
