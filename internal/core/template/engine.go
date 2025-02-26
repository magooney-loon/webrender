package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"github.com/magooney-loon/webserver/pkg/logger"
)

// New creates a new template engine
func New(log *logger.Logger, cfg Config) (*Engine, error) {
	e := &Engine{
		log:    log,
		cache:  make(map[string]*template.Template),
		config: cfg,
	}

	// Add common functions if not provided
	if len(cfg.CustomFuncs) == 0 {
		e.config.CustomFuncs = CommonFuncs()
	} else {
		// Merge common funcs with custom funcs
		commonFuncs := CommonFuncs()
		for name, fn := range commonFuncs {
			if _, exists := e.config.CustomFuncs[name]; !exists {
				e.config.CustomFuncs[name] = fn
			}
		}
	}

	if err := e.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return e, nil
}

// loadTemplates loads all templates from the configured directories
func (e *Engine) loadTemplates() error {
	e.log.Info("loading templates", map[string]interface{}{
		"templates_dir": e.config.TemplatesDir,
		"layouts_dir":   e.config.LayoutsDir,
		"partials_dir":  e.config.PartialsDir,
	})

	// Create base template with custom functions
	tmpl := template.New("").Funcs(e.config.CustomFuncs)

	// Set delimiters if custom ones are provided
	if e.config.DelimLeft != "" && e.config.DelimRight != "" {
		tmpl = tmpl.Delims(e.config.DelimLeft, e.config.DelimRight)
	}

	// First load layouts
	layoutsDir := filepath.Join(e.config.TemplatesDir, e.config.LayoutsDir)
	layoutFiles, err := filepath.Glob(filepath.Join(layoutsDir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to find layout files: %w", err)
	}

	e.log.Info("found layout files", map[string]interface{}{
		"layouts": layoutFiles,
	})

	// Then load partials
	partialsDir := filepath.Join(e.config.TemplatesDir, e.config.PartialsDir)

	// Load root partials
	rootPartials, err := filepath.Glob(filepath.Join(partialsDir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to find root partial files: %w", err)
	}

	// Load subdirectory partials
	subdirPartials, err := filepath.Glob(filepath.Join(partialsDir, "*/*.html"))
	if err != nil {
		return fmt.Errorf("failed to find partial files in subdirectories: %w", err)
	}

	// Combine all partials
	partialFiles := append(rootPartials, subdirPartials...)

	e.log.Info("found partial files", map[string]interface{}{
		"partials": partialFiles,
	})

	// Then load pages
	pagesDir := filepath.Join(e.config.TemplatesDir, "pages")
	pageFiles, err := filepath.Glob(filepath.Join(pagesDir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to find page files: %w", err)
	}

	e.log.Info("found page files", map[string]interface{}{
		"pages": pageFiles,
	})

	// Combine all template files
	templateFiles := append(layoutFiles, partialFiles...)
	templateFiles = append(templateFiles, pageFiles...)

	// Parse all templates
	tmpl, err = tmpl.ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("failed to parse template files: %w", err)
	}

	// Log all defined templates for debugging
	var definedTemplates []string
	for _, t := range tmpl.Templates() {
		definedTemplates = append(definedTemplates, t.Name())
	}

	e.log.Info("defined templates", map[string]interface{}{
		"templates": definedTemplates,
	})

	e.templates = tmpl
	return nil
}

// Render renders a template with the given data
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if e.config.ReloadOnRequest {
		if err := e.loadTemplates(); err != nil {
			return fmt.Errorf("error reloading templates: %w", err)
		}
	}

	// Check if template exists
	tmpl := e.templates.Lookup(name)

	// If not found, try with .html extension
	if tmpl == nil && !strings.HasSuffix(name, ".html") {
		nameWithExt := name + ".html"
		tmpl = e.templates.Lookup(nameWithExt)
		if tmpl != nil {
			// Update name for cache lookup
			name = nameWithExt
		}
	}

	// If still not found, return error with available templates
	if tmpl == nil {
		// Check for available templates
		var definedTemplates []string
		for _, t := range e.templates.Templates() {
			if t.Name() != "" {
				definedTemplates = append(definedTemplates, t.Name())
			}
		}
		return fmt.Errorf("template %q not found, available templates: %v", name, definedTemplates)
	}

	// Try to get template from cache first
	if e.config.CacheTemplates {
		e.cacheMutex.RLock()
		tmpl, exists := e.cache[name]
		e.cacheMutex.RUnlock()

		if exists {
			return tmpl.ExecuteTemplate(w, name, data)
		}
	}

	// Clone template to avoid race conditions
	tmpl, err := e.templates.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone templates: %w", err)
	}

	// Cache the template if enabled
	if e.config.CacheTemplates {
		e.cacheMutex.Lock()
		e.cache[name] = tmpl
		e.cacheMutex.Unlock()
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

// AddFunc adds a custom function to the template engine
func (e *Engine) AddFunc(name string, fn interface{}) {
	if e.config.CustomFuncs == nil {
		e.config.CustomFuncs = make(template.FuncMap)
	}
	e.config.CustomFuncs[name] = fn
}

// ReloadTemplates forces a reload of all templates
func (e *Engine) ReloadTemplates() error {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()

	e.cache = make(map[string]*template.Template)
	return e.loadTemplates()
}
