package user

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
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

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*GetMeResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*GetUserResponse, error)
	UpdateMe(ctx context.Context, user *UpdateUserRequest) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*GetMeResponse, error) {
	var res GetMeResponse

	builder := squirrel.Select(usernameColumn, firstNameColumn, bioColumn, avatarColumn, createdAtColumn).
		From(tableName).
		Where(squirrel.Eq{idColumn: id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, ErrCreateQuery
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&res.User.Username, &res.User.FirstName, &res.User.Bio, &res.User.Avatar, &res.User.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &res, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*GetUserResponse, error) {
	var res GetUserResponse

	builder := squirrel.Select(usernameColumn, firstNameColumn, bioColumn, avatarColumn, createdAtColumn).
		From(tableName).
		Where(squirrel.Eq{usernameColumn: username}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, ErrCreateQuery
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&res.User.Username, &res.User.FirstName, &res.User.Bio, &res.User.Avatar, &res.User.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &res, nil
}

func (r *userRepository) UpdateMe(ctx context.Context, req *UpdateUserRequest) error {
	builder := squirrel.Update(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{idColumn: req.Userid})

	if req.FirstName != nil {
		builder = builder.Set(firstNameColumn, *req.FirstName)
	}
	if req.Bio != nil {
		builder = builder.Set(bioColumn, *req.Bio)
	}
	if req.Avatar != nil {
		builder = builder.Set(avatarColumn, *req.Avatar)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return ErrCreateQuery
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
