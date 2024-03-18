package users

import (
	"context"
)

type ContextKey string

const ContextUserKey ContextKey = "context"

type User struct {
	ID       uint32 `json:"id"`
	Login    string `json:"username"`
	Role     string
	password string
}

type UserRepo interface {
	Authorize(login, pass string) (*User, error)
	Signup(login, pass string) (*User, error)
	UserExists(login string) (bool, error)
	GetUserRole(username string) (string, error)
	getUserByUsername(username string) (User, error)
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, ContextUserKey, user)
}
