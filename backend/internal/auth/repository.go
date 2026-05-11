package auth

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName       = "users"
	idColumn        = "id"
	usernameColumn  = "username"
	firstNameColumn = "first_name"
	passwordColumn  = "password"
	bioColumn       = "bio"
	avatarColumn    = "avatar"
	createdAtColumn = "created_at"
)

type AuthRepository interface {
	Login(ctx context.Context, req *LoginRequest) (*uuid.UUID, string, error)
	Register(ctx context.Context, req *RegisterRequest) (*uuid.UUID, error)
}

type authRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) Login(ctx context.Context, req *LoginRequest) (*uuid.UUID, string, error) {
	var id uuid.UUID
	var password string

	builder := squirrel.Select(idColumn, passwordColumn).From(tableName).Where(squirrel.Eq{usernameColumn: req.Username}).PlaceholderFormat(squirrel.Dollar)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", ErrCreateQuery
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&id, &password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrUserNotFound
		}

		return nil, "", ErrGetFromDB
	}

	return &id, password, nil
}

func (r *authRepository) Register(ctx context.Context, req *RegisterRequest) (*uuid.UUID, error) {
	var pgErr *pgconn.PgError
	id := uuid.New()

	builder := squirrel.Insert(tableName).Columns(idColumn, usernameColumn, firstNameColumn, passwordColumn).
		Values(id, req.Username, req.FirstName, req.Password).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, ErrCreateQuery
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, ErrUserAlreadyExists
			}
		}

		return nil, ErrInsertToDB
	}

	return &id, nil
}
