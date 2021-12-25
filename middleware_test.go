package pordego_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/miquella/pordego"
)

func TestMiddleware(t *testing.T) {
	c := qt.New(t)
	m := pordego.Middleware{
		Next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := pordego.UserFromContext(r.Context())
			c.Logf("User: %v", user)
			c.Fail()
		}),
	}
	r, _ := http.NewRequest("GET", "example.com/auth", nil)
	m.ServeHTTP(httptest.NewRecorder(), r)
}
