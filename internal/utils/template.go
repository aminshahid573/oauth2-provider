// File: internal/utils/template.go (Corrected)
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
	CSRFField template.HTML
	CSRFToken string
	Data      map[string]any
}

// TemplateCache caches parsed HTML templates.
type TemplateCache map[string]*template.Template

// NewTemplateCache parses all templates from the embedded filesystem and creates a cache.
func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}

	// Find all "page" templates (e.g., login.html, dashboard.html).
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
// It now accepts a layout file name.
func (tc TemplateCache) Render(w http.ResponseWriter, r *http.Request, layout, page string, data map[string]any) {
	ts, ok := tc[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		HandleError(w, r, slog.Default(), err)
		return
	}

	buf := new(bytes.Buffer)

	templateData := &TemplateData{
		CSRFField: csrf.TemplateField(r),
		CSRFToken: csrf.Token(r),
		Data:      data,
	}

	// Execute the specified layout template.
	err := ts.ExecuteTemplate(buf, layout, templateData)
	if err != nil {
		HandleError(w, r, slog.Default(), err)
		return
	}

	buf.WriteTo(w)
}
