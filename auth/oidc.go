package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenSet captures the fields we care about from an OIDC password-grant or
// refresh-token response. ObtainedAt is set client-side when the token is
// received so callers can compute true remaining lifetime.
type TokenSet struct {
	AccessToken      string    `json:"accessToken"`
	RefreshToken     string    `json:"refreshToken"`
	ExpiresIn        int       `json:"expiresIn"`        // seconds
	RefreshExpiresIn int       `json:"refreshExpiresIn"` // seconds
	TokenType        string    `json:"tokenType"`
	Scope            string    `json:"scope"`
	ObtainedAt       time.Time `json:"obtainedAt"`
}

// AccessExpiresAt is the wall-clock instant the access token expires.
func (t TokenSet) AccessExpiresAt() time.Time {
	return t.ObtainedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
}

// RefreshExpiresAt is the wall-clock instant the refresh token expires (or zero
// time if the server didn't tell us).
func (t TokenSet) RefreshExpiresAt() time.Time {
	if t.RefreshExpiresIn <= 0 {
		return time.Time{}
	}
	return t.ObtainedAt.Add(time.Duration(t.RefreshExpiresIn) * time.Second)
}

// ParseClaims decodes the payload segment of a JWT access token and returns
// it as a generic map. It does not verify the signature — callers use the
// claims for read-only purposes (display, sample-request scaffolding) and
// trust the token's issuer through the OIDC flow that produced it.
func ParseClaims(accessToken string) (map[string]any, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}
	parts := strings.Split(accessToken, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("access token is not a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding JWT payload: %w", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parsing JWT claims: %w", err)
	}
	return claims, nil
}

// raw is the minimal subset of the OIDC token response we parse.
type raw struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Login performs an OIDC Resource-Owner-Password-Credentials grant.
func Login(tokenEndpoint, clientID, username, password string) (TokenSet, error) {
	if tokenEndpoint == "" {
		return TokenSet{}, fmt.Errorf("token endpoint is not configured")
	}
	if clientID == "" {
		return TokenSet{}, fmt.Errorf("OIDC client_id is not configured")
	}
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", clientID)
	form.Set("username", username)
	form.Set("password", password)
	form.Set("scope", "openid")
	return post(tokenEndpoint, form)
}

// ResolveRealm fetches resolverURL (with {subdomain} substituted) and extracts
// the realm via dot-separated jsonPath. Used to derive a Keycloak realm name
// from a tenant subdomain before the password grant.
func ResolveRealm(resolverURL, subdomain, jsonPath string) (string, error) {
	if resolverURL == "" {
		return "", fmt.Errorf("realm resolver URL is empty")
	}
	if jsonPath == "" {
		return "", fmt.Errorf("realm JSON path is empty")
	}
	url := strings.ReplaceAll(resolverURL, "{subdomain}", subdomain)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("building resolver request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling realm resolver: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading resolver response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return "", fmt.Errorf("realm resolver returned status %d: %s", resp.StatusCode, snippet)
	}

	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return "", fmt.Errorf("parsing resolver JSON: %w", err)
	}
	val, ok := lookupJSONPath(doc, strings.Split(jsonPath, "."))
	if !ok {
		return "", fmt.Errorf("realm not found at path %q in resolver response", jsonPath)
	}
	s, ok := val.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("realm at path %q is empty or not a string", jsonPath)
	}
	return s, nil
}

func lookupJSONPath(doc any, path []string) (any, bool) {
	cur := doc
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, present := m[key]
		if !present {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

// Refresh exchanges a refresh token for a fresh TokenSet.
func Refresh(tokenEndpoint, clientID, refreshToken string) (TokenSet, error) {
	if tokenEndpoint == "" {
		return TokenSet{}, fmt.Errorf("token endpoint is not configured")
	}
	if clientID == "" {
		return TokenSet{}, fmt.Errorf("OIDC client_id is not configured")
	}
	if refreshToken == "" {
		return TokenSet{}, fmt.Errorf("refresh token is empty")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", clientID)
	form.Set("refresh_token", refreshToken)
	return post(tokenEndpoint, form)
}

func post(endpoint string, form url.Values) (TokenSet, error) {
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("calling token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenSet{}, fmt.Errorf("reading token response: %w", err)
	}

	var r raw
	if err := json.Unmarshal(body, &r); err != nil {
		// Body might be HTML/text on auth-server hiccups; surface a snippet.
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return TokenSet{}, fmt.Errorf("parsing token response (status=%d): %v; body=%s", resp.StatusCode, err, snippet)
	}

	if resp.StatusCode != http.StatusOK || r.AccessToken == "" {
		if r.Error != "" {
			return TokenSet{}, fmt.Errorf("oidc error: %s: %s", r.Error, r.ErrorDescription)
		}
		return TokenSet{}, fmt.Errorf("oidc returned status %d", resp.StatusCode)
	}

	return TokenSet{
		AccessToken:      r.AccessToken,
		RefreshToken:     r.RefreshToken,
		ExpiresIn:        r.ExpiresIn,
		RefreshExpiresIn: r.RefreshExpiresIn,
		TokenType:        r.TokenType,
		Scope:            r.Scope,
		ObtainedAt:       time.Now(),
	}, nil
}
