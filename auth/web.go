package auth

import (
	"context"
	"net/http"
)

func InterceptBasicAuth(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inUsername, inPw, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx := context.WithValue(r.Context(), "username", inUsername)
			ctx = context.WithValue(ctx, "password", inPw)
			delegate(w, r.WithContext(ctx))
		}
	}
}

func InterceptAuth(tokenManager *TokenManager, delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenEntry := validateToken(r, tokenManager)
		if tokenEntry == nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx := context.WithValue(r.Context(), "tokenEntry", tokenEntry)
			delegate(w, r.WithContext(ctx))
		}
	}
}

func validateToken(r *http.Request, tokenManager *TokenManager) *TokenEntry {
	token, ok := resolveToken(r)
	if ok {
		return tokenManager.ValidateToken(token)
	} else {
		return nil
	}
}

func resolveToken(r *http.Request) (string, bool) {
	c, err := r.Cookie("authToken")
	if err == nil {
		return c.Value, true
	}
	value := r.Header.Get("authToken")
	if len(value) > 0 {
		return value, true
	}
	values, ok := r.URL.Query()["authToken"]
	if ok {
		return values[0], true
	}
	return "", false
}
