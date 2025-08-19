package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/http"

	"github.com/aminshahid573/oauth2-provider/internal/store"
)

func AuthorizeGetHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")

	if !store.IsValidClient(clientID, redirectURI) || responseType != "code" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_user")

	if err != nil || cookie.Value == "" {
		// Redirect back to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/consent,html"))
	tmpl.Execute(w, map[string]string{
		"ClientID":    clientID,
		"RedirectURI": redirectURI,
	})
}

func AuthorizePostHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")

	if !store.IsValidClient(clientID, redirectURI) || responseType != "code" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if user is logged in
	cookie, err := r.Cookie("session_user")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user approved consent (you might want to check a form field here)
	if r.FormValue("approve") != "true" {
		// User denied consent
		http.Redirect(w, r, "/login", http.StatusBadRequest)
		return
	}

	// Generate and save authorization code
	code := generateCode()
	store.SaveCode(code, cookie.Value)

	// Redirect back to client with authorization code
	http.Redirect(w, r, redirectURI+"?code="+code, http.StatusFound)
}

func generateCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
