package auth

import (
	"context"
	"pooky-messanger/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (string, error)
}

type authService struct {
	r AuthRepository
	t jwt.JWTService
}

func NewAuthService(r AuthRepository, t jwt.JWTService) AuthService {
	return &authService{
		r: r,
		t: t,
	}
}

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (string, error) {
	if len(req.Username) < 4 || len(req.Password) < 8 || len(req.FirstName) < 2 {
		return "", ErrInvalidParams
	}

	hashPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	user := &RegisterRequest{
		Username:  req.Username,
		Password:  string(hashPassword),
		FirstName: req.FirstName,
	}

	res, err := s.r.Register(ctx, user)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	var res AuthResponse

	if len(req.Username) < 4 || len(req.Password) < 8 {
		return nil, ErrInvalidParams
	}

	id, password, err := s.r.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password))
	if err != nil {
		return nil, ErrLoginUser
	}

	token, err := s.t.GetToken(id.String())
	if err != nil {
		return nil, err
	}

	res.Token = *token

	return &res, nil
}
