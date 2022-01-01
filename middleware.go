package pordego

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	StateCookie = "_state"
)

type Middleware struct {
	Config   oauth2.Config
	Verifier *oidc.IDTokenVerifier

	Next http.Handler
}

var _ (http.Handler) = (*Middleware)(nil)

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	if !hasOIDCResponse(r) {
		err = m.startAuth(w, r)
	} else {
		err = m.completeAuth(w, r)
	}

	if err != nil {
		log.Printf("[Pordego] ERROR: %s", err.Error())

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(resperr.StatusCode(err))
		fmt.Fprint(w, resperr.UserMessage(err))
	}
}

func hasOIDCResponse(r *http.Request) bool {
	query := r.URL.Query()
	return query.Has("code") || query.Has("error")
}

func (m *Middleware) startAuth(w http.ResponseWriter, r *http.Request) error {
	// Set the state cookie
	state, err := generateState()
	if err != nil {
		// TODO: Handle the error
	}

	http.SetCookie(w, &http.Cookie{
		Name:     StateCookie,
		Value:    state,
		Path:     r.URL.Path,
		HttpOnly: true,
	})

	// Redirect the client
	redirectURL := m.Config.AuthCodeURL(state)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

	return nil
}

func (m *Middleware) completeAuth(w http.ResponseWriter, r *http.Request) error {
	// TODO: Handle r.FormValue("error")

	// Validate state
	stateCookie, err := r.Cookie(StateCookie)
	if err != nil {
		return resperr.WithCodeAndMessage(nil, http.StatusBadRequest, "State not found. Unable to complete authentication.")
	}

	if stateCookie.Value != r.FormValue("state") {
		return resperr.WithCodeAndMessage(nil, http.StatusBadRequest, "State mismatch. Unable to complete authentication.")
	}

	// TODO: Now that the state has been verified, we should delete the cookie

	// Exchange code for a token
	// TODO: Should PKCE be used here?
	token, err := m.Config.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		return resperr.WithCodeAndMessage(err, http.StatusForbidden, "OAuth token exchange failed. Unable to complete authentication.")
	}

	// Verify the OIDC token
	if m.Verifier == nil {
		return resperr.WithCodeAndMessage(nil, http.StatusInternalServerError, "Verifier not configured. Unable to verify OIDC token")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return resperr.WithCodeAndMessage(nil, http.StatusForbidden, "Invalid OIDC token. Unable to complete authentication.")
	}

	idToken, err := m.Verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		return resperr.WithCodeAndMessage(err, http.StatusForbidden, "OIDC token verification failed. Unable to complete authentication.")
	}

	// Pass the user to the next middleware
	var claims struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	_ = idToken.Claims(&claims)

	if m.Next != nil {
		user := User{
			ID:    idToken.Subject,
			Name:  claims.Name,
			Email: claims.Email,
		}

		userCtx := ContextWithUser(r.Context(), &user)
		m.Next.ServeHTTP(w, r.WithContext(userCtx))
	}

	return nil
}

func generateState() (string, error) {
	state := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, state)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(state), nil
}
