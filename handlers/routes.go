package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"who-owes-me/auth"
	"who-owes-me/db"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

func RegisterRoutes(r chi.Router) {
	r.Get("/health", handleHealth)
	r.Get("/login", handleLogin)
	r.Get("/callback", handleCallback)
	r.Get("/logout", handleLogout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Get("/", handleDashboard)
		r.Get("/users/{sub}", handleUserDashboardBySub)

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(AdminMiddleware)
			r.Get("/admin", handleAdminDashboard)
			r.Post("/admin/users", handleCreateUser)
			r.Post("/admin/users/update", handleUpdateUser)
			r.Post("/admin/splits", handleCreateSplits)
			r.Get("/admin/payees", handleGetPayees) // HTMX endpoint
		})
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if auth.Provider == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	auth.SetCookie(w, "oauth_state", state)

	url := auth.OAuth2Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GetCookie(r, "oauth_state")
	if err != nil || r.URL.Query().Get("state") != state {
		http.Error(w, "State invalid", http.StatusBadRequest)
		return
	}

	ctx := auth.GetClientContext(r.Context())
	oauth2Token, err := auth.OAuth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
		return
	}

	_, err = auth.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch UserInfo to get groups since Authelia 4.39 removes groups from the ID Token by default
	userInfo, err := auth.Provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to fetch user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var claims auth.CustomClaims
	if err := userInfo.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse user info claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	isAdmin := false
	for _, group := range claims.Groups {
		if group == "whoowesme_admin" {
			isAdmin = true
			break
		}
	}

	auth.SetAdminCookie(w, isAdmin)
	auth.SetCookie(w, "auth_token", rawIDToken)
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	idToken, _ := auth.GetCookie(r, "auth_token")

	auth.ClearCookie(w, "auth_token")
	auth.ClearCookie(w, "is_admin")

	if auth.Provider != nil {
		var providerClaims struct {
			EndSessionEndpoint string `json:"end_session_endpoint"`
		}
		if err := auth.Provider.Claims(&providerClaims); err == nil && providerClaims.EndSessionEndpoint != "" {
			redirectURL := providerClaims.EndSessionEndpoint
			if idToken != "" {
				redirectURL += "?id_token_hint=" + idToken
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

// Middleware

type contextKey string

const userCtxKey = contextKey("user")
const isAdminCtxKey = contextKey("isAdmin")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In development mode or when OIDC is not configured, bypass authentication
		if auth.Provider == nil {
			user, err := db.GetUserBySub("dev_user")
			if err != nil {
				db.CreateUser("Dev User", "dev_user", "regular", "dev_payee")
				user, _ = db.GetUserBySub("dev_user")
			}

			ctx := context.WithValue(r.Context(), userCtxKey, user)
			ctx = context.WithValue(ctx, isAdminCtxKey, true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		tokenStr, err := auth.GetCookie(r, "auth_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		authCtx := auth.GetClientContext(r.Context())
		idToken, err := auth.Verifier.Verify(authCtx, tokenStr)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		isAdmin, err := auth.GetAdminCookie(r)
		if err != nil {
			// Invalid or missing cookie (e.g. app restarted), force re-login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Lookup user in DB
		user, err := db.GetUserBySub(idToken.Subject)
		if err != nil {
			if isAdmin {
				// Admins are allowed to proceed even if not in DB, to bootstrap
				ctx := context.WithValue(r.Context(), isAdminCtxKey, true)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			w.WriteHeader(http.StatusForbidden)
			renderTemplate(w, "unregistered.html", nil)
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		ctx = context.WithValue(ctx, isAdminCtxKey, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value(isAdminCtxKey).(bool)
		if !ok || !isAdmin {
			renderError(w, http.StatusForbidden, "Admin access required.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
