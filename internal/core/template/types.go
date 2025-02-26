package template

import (
	"html/template"
	"sync"

	"github.com/magooney-loon/webserver/pkg/logger"
)

// Engine represents the template engine
type Engine struct {
	templates  *template.Template
	log        *logger.Logger
	cache      map[string]*template.Template
	cacheMutex sync.RWMutex
	config     Config
}

// Config holds template engine configuration
type Config struct {
	TemplatesDir    string
	LayoutsDir      string
	PartialsDir     string
	DefaultLayout   string
	CacheTemplates  bool
	DelimLeft       string
	DelimRight      string
	CustomFuncs     template.FuncMap
	ReloadOnRequest bool
	Development     bool
}

// DefaultConfig returns default template engine configuration
func DefaultConfig() Config {
	return Config{
		TemplatesDir:    "web/templates",
		LayoutsDir:      "layouts",
		PartialsDir:     "partials",
		DefaultLayout:   "base",
		CacheTemplates:  true,
		DelimLeft:       "{{",
		DelimRight:      "}}",
		CustomFuncs:     make(template.FuncMap),
		ReloadOnRequest: false,
		Development:     false,
	}
}
