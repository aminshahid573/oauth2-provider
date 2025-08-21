package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/aminshahid573/oauth2-provider/web"
	"github.com/gorilla/csrf"
)

// TemplateData holds data to be passed to HTML templates.
type TemplateData struct {
	// CSRFField is the hidden input field for CSRF protection.
	CSRFField template.HTML
	// Add other common fields here, e.g., IsAuthenticated, User, Flash messages.
	Data map[string]any
}

// TemplateCache caches parsed HTML templates.
type TemplateCache map[string]*template.Template

// NewTemplateCache parses all templates from the embedded filesystem and creates a cache.
func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}

	// We want to find all "page" templates (e.g., login.html, consent.html).
	// These are the entry points for rendering.
	pages, err := fs.Glob(web.Templates, "templates/auth/*.html")
	if err != nil {
		return nil, err
	}
	pagesAdmin, err := fs.Glob(web.Templates, "templates/admin/*.html")
	if err != nil {
		return nil, err
	}
	pages = append(pages, pagesAdmin...)

	for _, page := range pages {
		name := filepath.Base(page)

		// Create a new template set for each page.
		// Add our custom functions.
		ts, err := template.New(name).ParseFS(web.Templates, page)
		if err != nil {
			return nil, err
		}

		// Parse layout templates and add them to the set.
		ts, err = ts.ParseFS(web.Templates, "templates/layouts/*.html")
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

// Render executes a template from the cache and writes it to the response.
func (tc TemplateCache) Render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	ts, ok := tc[name]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", name)
		HandleError(w, r, slog.Default(), err)
		return
	}

	buf := new(bytes.Buffer)

	templateData := &TemplateData{
		CSRFField: csrf.TemplateField(r),
		Data:      data,
	}

	// Execute the "base.html" template. It will pull in the blocks defined
	// in the page template (e.g., "login.html") because they are in the same template set.
	err := ts.ExecuteTemplate(buf, "base.html", templateData)
	if err != nil {
		HandleError(w, r, slog.Default(), err)
		return
	}

	buf.WriteTo(w)
}
