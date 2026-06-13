package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	Provider     *oidc.Provider
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
)

func InitOIDC() error {
	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := os.Getenv("OIDC_REDIRECT_URL")

	if issuerURL == "" || clientID == "" {
		return fmt.Errorf("OIDC configuration is missing")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if os.Getenv("DOCKER_ENV") == "true" {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if addr == "localhost:9091" {
				addr = "authelia:9091"
			}
			return net.Dial(network, addr)
		}
	}
	client := &http.Client{Transport: transport}
	ctx := oidc.ClientContext(context.Background(), client)

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return fmt.Errorf("failed to get provider: %v", err)
	}

	Provider = provider
	Verifier = provider.Verifier(&oidc.Config{ClientID: clientID})
	OAuth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	return nil
}

type CustomClaims struct {
	Groups []string `json:"groups"`
	Email  string   `json:"email"`
}

// Helpers for cookies
func SetCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	})
}

func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
