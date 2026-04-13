package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	FirebaseAPIKey   = "AIzaSyD2Sd_NP3vA4rwvoroKqDefpXZeCMDXcIQ"
	FirebaseTenantID = "PRODTwinMind-dcnoy"
	GoogleClientID   = "352176597832-pjhsenkctronnke9u70d0pvhanmscc2hr.apps.googleusercontent.com"

	tokenURL      = "https://securetoken.googleapis.com/v1/token"
	signInWithIdp = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithIdp"
	googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
)

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

type signInWithIdpRequest struct {
	RequestURI        string `json:"requestUri"`
	PostBody          string `json:"postBody"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
	TenantID          string `json:"tenantId"`
}

type signInResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

type googleTokenResponse struct {
	IDToken string `json:"id_token"`
}

func ExchangeRefreshToken(refreshToken string) (*TokenPair, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	resp, err := http.Post(
		tokenURL+"?key="+FirebaseAPIKey,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	var result refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresIn, _ := time.ParseDuration(result.ExpiresIn + "s")

	return &TokenPair{
		IDToken:      result.IDToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    time.Now().Add(expiresIn),
	}, nil
}

// ExchangeGoogleCode exchanges a Google OAuth authorization code for a Google ID token.
func ExchangeGoogleCode(code, redirectURI string) (string, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {GoogleClientID},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.Post(
		googleTokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("google token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode google token response: %w", err)
	}

	return result.IDToken, nil
}

// SignInWithGoogle takes a Google ID token and exchanges it for Firebase credentials.
func SignInWithGoogle(googleIDToken string) (*TokenPair, error) {
	postBody := "&id_token=" + url.QueryEscape(googleIDToken) + "&providerId=google.com"

	req := signInWithIdpRequest{
		RequestURI:        "http://localhost",
		PostBody:          postBody,
		ReturnSecureToken: true,
		TenantID:          FirebaseTenantID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		signInWithIdp+"?key="+FirebaseAPIKey,
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("firebase sign in failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("firebase sign in failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result signInResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode sign in response: %w", err)
	}

	expiresIn, _ := time.ParseDuration(result.ExpiresIn + "s")

	return &TokenPair{
		IDToken:      result.IDToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    time.Now().Add(expiresIn),
	}, nil
}
