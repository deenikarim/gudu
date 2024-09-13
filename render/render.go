package render

import (
	"bytes"
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

type Render struct {
	RendererEngine    string
	TemplatesRootPath string
	Secure            bool
	Port              string
	ServerName        string
	JetViews          *jet.Set
	GoTemplateCache   sync.Map
	CustomsFuncs      template.FuncMap
	Session           *scs.SessionManager
	// DefaultData       *TemplateData
	DevelopmentMode bool
	once            sync.Once
}

type TemplateData struct {
	IsUserAuthenticated bool
	IntMap              map[string]int
	StringMap           map[string]string
	FloatMap            map[string]float64
	GenericData         map[string]any
	CSRFToken           string
	Secure              bool
	Port                string
	ServerName          string
	FormData            url.Values
	Errors              map[string][]string
}

// AddDefaultsData add common dynamic data on every webpage
func (r *Render) AddDefaultsData(td *TemplateData, rr *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}

	td.ServerName = r.ServerName
	td.Port = r.Port
	td.Secure = r.Secure
	if r.Session.Exists(rr.Context(), "user_id") {
		td.IsUserAuthenticated = true
	}

	return td
}

// AddCustomFuncs adds a custom template function to the Render instance.
func (r *Render) AddCustomFuncs(funcMaps template.FuncMap) {
	if r.CustomsFuncs == nil {
		r.CustomsFuncs = make(template.FuncMap)
	}
	for name, fn := range funcMaps {
		r.CustomsFuncs[name] = fn
	}
}

// RenderPage specifies default template rendering engine
func (r *Render) RenderPage(w http.ResponseWriter, rr *http.Request, templateName string, variables, data any) error {
	switch strings.ToLower(r.RendererEngine) {
	case "go":
		return r.RenderGoPage(w, rr, templateName, data)
	case "jet":
		return r.RenderJetPage(w, rr, templateName, variables, data)

	}
	return nil
}

// todo: jet template engine support

// RenderJetPage render jet template engine
func (r *Render) RenderJetPage(w http.ResponseWriter, rr *http.Request, templateName string, variables, data any) error {
	var varsData jet.VarMap
	if variables == nil {
		varsData = make(jet.VarMap)
	} else {
		varsData = variables.(jet.VarMap)
	}
	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = r.AddDefaultsData(td, rr)

	t, err := r.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println(err)
		return err
	}
	if err := t.Execute(w, varsData, &td); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// todo: Go template engine support

// ParseTemplates parses all templates in the directory and cache as map.
func (r *Render) ParseTemplates() error {
	// layouts template
	layoutFiles, err := filepath.Glob(filepath.Join("views", "layouts/*layout.gohtml"))
	if err != nil {
		return fmt.Errorf("error globbing layout files: %v", err)
	}

	// get pages template
	pageFiles, err := filepath.Glob(filepath.Join("views", "pages/*.gohtml"))
	if err != nil {
		return fmt.Errorf("error globbing files files: %v", err)
	}

	for _, page := range pageFiles {
		files := append(layoutFiles, page)
		name := filepath.Base(page)
		tmpl, err := template.New(name).Funcs(r.CustomsFuncs).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("error parsing template files: %v", err)
		}

		r.GoTemplateCache.Store(name, tmpl)
	}
	log.Println("Parsed and cached templates")
	return nil
}

// cacheTemplates ensures templates are cached once in production mode.
func (r *Render) cacheTemplates() {
	// Ensures the function inside is executed only once
	r.once.Do(func() {
		if err := r.ParseTemplates(); err != nil {
			log.Printf("Failed to load and cache templates: %v\n", err)
		} else {
			log.Println("Templates cached successfully.")
		}
	})
}

// RenderGoPage retrieves the specified template from the cache or loads it if in development mode and then executes it.
func (r *Render) RenderGoPage(w http.ResponseWriter, rr *http.Request, templateName string, data any) error {
	// Load or cache templates as needed
	if r.DevelopmentMode {
		// Reload templates on each request in development mode
		if err := r.ParseTemplates(); err != nil {
			log.Printf("error parsing templates: %v\n", err)
			http.Error(w, "Error parsing templates.", http.StatusInternalServerError)
			return err
		}
	} else {
		// Ensure templates are cached only once in production mode
		r.cacheTemplates()
	}

	// Retrieve the template from the cache
	tmpl, ok := r.GoTemplateCache.Load(templateName)

	if !ok {
		return fmt.Errorf("template %s does not exist", templateName)
	}

	// check if data is passed to the template
	var td *TemplateData
	if data != nil {
		var ok bool
		td, ok = data.(*TemplateData)
		if !ok {
			http.Error(w, "Invalid template data.", http.StatusInternalServerError)
			return nil
		}
	}

	// add default data
	td = r.AddDefaultsData(td, rr)

	// Execute the template
	buf := new(bytes.Buffer)
	if err := tmpl.(*template.Template).Execute(buf, td); err != nil {
		log.Printf("error executing template to buffer: %v\n", err)
		http.Error(w, "Error buffer template.", http.StatusInternalServerError)
		return err
	}

	// Write the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("error writing template to the browser: %v\n", err)
		http.Error(w, "Error rendering template.", http.StatusInternalServerError)
		return err
	}
	return nil
}
