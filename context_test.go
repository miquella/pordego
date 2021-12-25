package pordego_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/miquella/pordego"
)

func TestContextWithUser(t *testing.T) {
	c := qt.New(t)
	user := pordego.User{ID: "an-id", Email: "an-id@an-email.zzz"}
	ctx := pordego.ContextWithUser(context.Background(), &user)
	c.Assert(pordego.UserFromContext(ctx), qt.DeepEquals, &user)
}

func TestUserFromContext_NoUserFromBackgroundContext(t *testing.T) {
	c := qt.New(t)
	user := pordego.UserFromContext(context.Background())
	c.Assert(user, qt.IsNil)
}
