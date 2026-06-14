package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/http"

	"who-owes-me/auth"

	"who-owes-me/db"
	"who-owes-me/internal/envutil"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
)

// Helper for local docker OIDC mapping
func clientContext(ctx context.Context) context.Context {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	if envutil.Getenv("DOCKER_ENV") == "true" {
		customTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if addr == "localhost:9091" {
				addr = "authelia:9091"
			}
			return net.Dial(network, addr)
		}
	}
	return oidc.ClientContext(ctx, &http.Client{Transport: customTransport})
}

func RegisterRoutes(r chi.Router) {
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

	ctx := clientContext(r.Context())
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

	idToken, err := auth.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var claims auth.CustomClaims
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	auth.SetCookie(w, "auth_token", rawIDToken)
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearCookie(w, "auth_token")
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

		authCtx := clientContext(r.Context())
		idToken, err := auth.Verifier.Verify(authCtx, tokenStr)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		var claims auth.CustomClaims
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		// Check if admin
		isAdmin := false
		for _, group := range claims.Groups {
			if group == "admin" || group == "admins" { // adjust based on authelia config
				isAdmin = true
				break
			}
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
