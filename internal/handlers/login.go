package handlers

import (
	"html/template"
	"net/http"
)

func LoginGetHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("internal/templates/login.html"))
	tmpl.Execute(w, nil)

}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "admin" && password == "password" {
		http.SetCookie(w, &http.Cookie{
			Name:  "session_user",
			Value: username,
			Path:  "/",
		})
		http.Redirect(w, r, "/authorize", http.StatusSeeOther)
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}
