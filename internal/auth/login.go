package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/browser"
)

const (
	callbackPath = "/callback"
	loginTimeout = 120 * time.Second
	startPort    = 8742
)

// loginPage serves an HTML page that uses the Firebase Auth SDK to perform
// sign-in client-side via Google or Apple, then POSTs the resulting tokens
// back to the local server.
const loginPage = `<!DOCTYPE html>
<html>
<head><title>TwinMind - Sign In</title></head>
<body style="font-family:system-ui;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#f8f9fa">
<div id="status" style="text-align:center">
<h1 style="margin-bottom:8px">TwinMind</h1>
<p style="color:#666;margin-bottom:24px">Choose a sign-in method</p>
<button id="google-btn" style="display:flex;align-items:center;justify-content:center;gap:10px;width:280px;padding:12px 16px;margin:0 auto 12px;border:1px solid #ddd;border-radius:8px;background:#fff;font-size:15px;cursor:pointer;font-family:system-ui">
<svg width="20" height="20" viewBox="0 0 48 48"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/><path fill="#FBBC05" d="M10.53 28.59a14.5 14.5 0 010-9.18l-7.98-6.19a24.003 24.003 0 000 21.56l7.98-6.19z"/><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/></svg>
Sign in with Google
</button>
<button id="apple-btn" style="display:flex;align-items:center;justify-content:center;gap:10px;width:280px;padding:12px 16px;margin:0 auto;border:none;border-radius:8px;background:#000;color:#fff;font-size:15px;cursor:pointer;font-family:system-ui">
<svg width="20" height="20" viewBox="0 0 24 24" fill="#fff"><path d="M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.48-3.24 0-1.44.62-2.2.44-3.06-.4C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/></svg>
Sign in with Apple
</button>
</div>
<script src="https://www.gstatic.com/firebasejs/10.12.2/firebase-app-compat.js"></script>
<script src="https://www.gstatic.com/firebasejs/10.12.2/firebase-auth-compat.js"></script>
<script>
firebase.initializeApp({
  apiKey: "` + FirebaseAPIKey + `",
  authDomain: "thirdear-ai.firebaseapp.com"
});

var auth = firebase.auth();
auth.tenantId = "` + FirebaseTenantID + `";

function sendTokens(result) {
  return result.user.getIdTokenResult().then(function(tokenResult) {
    return fetch("/callback", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        id_token: tokenResult.token,
        refresh_token: result.user.refreshToken,
        expiration: tokenResult.expirationTime
      })
    });
  }).then(function() {
    document.getElementById("status").innerHTML = '<h1 style="color:#22c55e">Authentication Successful</h1><p>You can close this window and return to the terminal.</p>';
  });
}

function showError(err) {
  if (err.code === "auth/popup-closed-by-user") return;
  document.getElementById("status").innerHTML = '<h1 style="color:#ef4444">Authentication Failed</h1><p>' + err.message + '</p>';
}

document.getElementById("google-btn").addEventListener("click", function() {
  var provider = new firebase.auth.GoogleAuthProvider();
  provider.addScope("email");
  provider.addScope("profile");
  provider.setCustomParameters({ prompt: "select_account" });
  auth.signInWithPopup(provider).then(sendTokens).catch(showError);
});

document.getElementById("apple-btn").addEventListener("click", function() {
  var provider = new firebase.auth.OAuthProvider("apple.com");
  provider.addScope("email");
  provider.addScope("name");
  auth.signInWithPopup(provider).then(sendTokens).catch(showError);
});
</script>
</body>
</html>`

const successHTML = `<!DOCTYPE html>
<html>
<head><title>TwinMind - Authenticated</title></head>
<body style="font-family:system-ui;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#f8f9fa">
<div style="text-align:center">
<h1 style="color:#22c55e">Authentication Successful</h1>
<p>You can close this window and return to the terminal.</p>
</div>
</body>
</html>`

type callbackPayload struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Expiration   string `json:"expiration"`
}

func findFreePort() (int, error) {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", startPort, startPort+100)
}

func Login() (*TokenPair, error) {
	port, err := findFreePort()
	if err != nil {
		return nil, err
	}

	resultCh := make(chan *TokenPair, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()

	// Serve the login page with Firebase Auth SDK
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, loginPage)
	})

	// Receive tokens from the Firebase Auth SDK via POST
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload callbackPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			errCh <- fmt.Errorf("failed to decode callback payload: %w", err)
			return
		}

		if payload.IDToken == "" {
			http.Error(w, "missing id_token", http.StatusBadRequest)
			errCh <- fmt.Errorf("callback missing id_token")
			return
		}

		expiresAt, _ := time.Parse(time.RFC3339, payload.Expiration)
		if expiresAt.IsZero() {
			// Default to 1 hour from now if parsing fails
			expiresAt = time.Now().Add(1 * time.Hour)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

		resultCh <- &TokenPair{
			IDToken:      payload.IDToken,
			RefreshToken: payload.RefreshToken,
			ExpiresAt:    expiresAt,
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	loginURL := fmt.Sprintf("http://localhost:%d", port)
	fmt.Println("Opening browser for authentication...")
	if err := browser.OpenURL(loginURL); err != nil {
		fmt.Printf("Could not open browser. Please visit:\n%s\n", loginURL)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	select {
	case token := <-resultCh:
		return token, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(loginTimeout):
		return nil, fmt.Errorf("authentication timed out after %s", loginTimeout)
	}
}
