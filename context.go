package pordego

import (
	"context"
)

type userKey struct{}

type User struct {
	ID    string
	Name  string
	Email string
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userKey{}).(*User)
	return user
}
