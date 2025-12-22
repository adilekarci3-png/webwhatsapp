package user

import "context"

type Repository interface {
	Ensure(ctx context.Context, id string) error
}
