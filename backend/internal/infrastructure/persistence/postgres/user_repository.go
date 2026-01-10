package postgres

import (
	"context"

	"example.com/webwhatsapp/backend/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `
SELECT id, name, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;
`
	var u user.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	const q = `
SELECT id, name, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;
`
	var u user.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	const q = `
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, password_hash, created_at, updated_at;
`
	var out user.User
	err := r.pool.QueryRow(ctx, q, u.Name, u.Email, u.PasswordHash).Scan(
		&out.ID,
		&out.Name,
		&out.Email,
		&out.PasswordHash,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
