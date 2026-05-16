package user

import "time"

type User struct {
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	Bio       *string   `json:"bio"`
	Avatar    *string   `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}
type GetMeRequest struct {
	Userid string
}

type GetMeResponse struct {
	User User `json:"user"`
}

type UpdateUserRequest struct {
	Userid    string
	FirstName *string
	Bio       *string
	Avatar    *string
}

type UpdateUserResponse struct {
	User User `json:"user"`
}

type GetUserRequest struct {
	Userid   string
	Username string
}

type GetUserResponse struct {
	User User `json:"user"`
}
