package ports

import (
	"context"

	"example.com/webwhatsapp/backend/internal/domain/user"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
	Create(ctx context.Context, u *user.User) (*user.User, error)
}
