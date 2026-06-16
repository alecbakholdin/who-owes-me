package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"who-owes-me/internal/envutil"
)

var (
	Provider     *oidc.Provider
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
	cookieSecret = make([]byte, 32)
)

func init() {
	rand.Read(cookieSecret)
}

// httpsToHttpTransport rewrites HTTPS to HTTP for local testing against Authelia 4.38+
type httpsToHttpTransport struct {
	T http.RoundTripper
}

func (d *httpsToHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	if req2.URL.Scheme == "https" {
		req2.URL.Scheme = "http"
	}
	req2.Header.Set("X-Forwarded-Proto", "https")
	return d.T.RoundTrip(req2)
}

// GetClientContext creates an OIDC context with proper local docker mapping and HTTP overrides
func GetClientContext(ctx context.Context) context.Context {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	if envutil.Getenv("DOCKER_ENV") == "true" {
		customTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if addr == "localhost:9091" || addr == "authelia.localhost:9091" {
				addr = "authelia:9091"
			}
			return net.Dial(network, addr)
		}
	}

	issuerURL := envutil.Getenv("OIDC_ISSUER_URL")
	if strings.Contains(issuerURL, "localhost") {
		return oidc.ClientContext(ctx, &http.Client{
			Transport: &httpsToHttpTransport{T: customTransport},
		})
	}

	return oidc.ClientContext(ctx, &http.Client{Transport: customTransport})
}

func InitOIDC() error {
	issuerURL := envutil.Getenv("OIDC_ISSUER_URL")
	clientID := envutil.Getenv("OIDC_CLIENT_ID")
	clientSecret := envutil.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := envutil.Getenv("OIDC_REDIRECT_URL")

	if issuerURL == "" || clientID == "" {
		return fmt.Errorf("OIDC configuration is missing")
	}

	ctx := GetClientContext(context.Background())

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
		Secure:   envutil.Getenv("APP_ENV") == "production",
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
		Secure:   envutil.Getenv("APP_ENV") == "production",
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

// SetAdminCookie securely signs and sets the admin status
func SetAdminCookie(w http.ResponseWriter, isAdmin bool) {
	val := "false"
	if isAdmin {
		val = "true"
	}
	mac := hmac.New(sha256.New, cookieSecret)
	mac.Write([]byte(val))
	sig := hex.EncodeToString(mac.Sum(nil))
	SetCookie(w, "is_admin", val+"|"+sig)
}

// GetAdminCookie retrieves and verifies the admin status from the signed cookie
func GetAdminCookie(r *http.Request) (bool, error) {
	cookie, err := GetCookie(r, "is_admin")
	if err != nil {
		return false, err
	}
	parts := strings.Split(cookie, "|")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid cookie format")
	}
	val := parts[0]
	sig := parts[1]

	mac := hmac.New(sha256.New, cookieSecret)
	mac.Write([]byte(val))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(sig), []byte(expectedSig)) == 1 {
		return val == "true", nil
	}
	return false, fmt.Errorf("invalid cookie signature")
}
