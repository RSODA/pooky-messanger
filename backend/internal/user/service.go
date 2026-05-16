package user

import (
	"context"
	"pooky-messanger/pkg/storage"
)

type UserService interface {
	GetMe(ctx context.Context, req *GetMeRequest) (*GetMeResponse, error)
	GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)
	UpdateMe(ctx context.Context, req *UpdateUserRequest) error
}

type userService struct {
	r UserRepository
}

func NewUserService(r UserRepository) UserService {
	return &userService{r: r}
}

func (s *userService) GetMe(ctx context.Context, req *GetMeRequest) (*GetMeResponse, error) {
	if len(req.Userid) == 0 {
		return nil, ErrInvalidToken
	}

	res, err := s.r.GetUserByID(ctx, req.Userid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *userService) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	if len(req.Userid) == 0 {
		return nil, ErrInvalidToken
	}

	res, err := s.r.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *userService) UpdateMe(ctx context.Context, req *UpdateUserRequest) error {
	if len(req.Userid) == 0 {
		return ErrInvalidToken
	}

	if req.Avatar != nil {
		file, err := storage.Upload(*req.Avatar)
		if err != nil {
			return ErrUpdateAvatar
		}

		req.Avatar = &file
	}

	err := s.r.UpdateMe(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
