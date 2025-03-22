package template

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Manager handles template loading, caching, and rendering
type Manager struct {
	// Template storage
	templates     map[string]*template.Template
	templateMutex sync.RWMutex

	// Base templates
	baseTemplates     map[string]*template.Template
	baseTemplateMutex sync.RWMutex
}

// NewManager creates a new template manager
func NewManager() *Manager {
	return &Manager{
		templates:     make(map[string]*template.Template),
		baseTemplates: make(map[string]*template.Template),
	}
}

// RegisterTemplate registers a template with the manager
func (m *Manager) RegisterTemplate(name, filePath string) error {
	// Parse template
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", filePath, err)
	}

	// Store template
	m.templateMutex.Lock()
	defer m.templateMutex.Unlock()
	m.templates[name] = tmpl

	return nil
}

// RegisterBaseTemplate registers a base template
func (m *Manager) RegisterBaseTemplate(name, filePath string) error {
	// Parse template
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse base template %s: %w", filePath, err)
	}

	// Store base template
	m.baseTemplateMutex.Lock()
	defer m.baseTemplateMutex.Unlock()
	m.baseTemplates[name] = tmpl

	return nil
}

// RenderTemplate renders a template with the given data
func (m *Manager) RenderTemplate(w io.Writer, name, filePath string, data interface{}) error {
	// Check if template is cached
	m.templateMutex.RLock()
	tmpl, exists := m.templates[name]
	m.templateMutex.RUnlock()

	// Parse template if not cached
	if !exists {
		var err error
		tmpl, err = template.ParseFiles(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", filePath, err)
		}

		// Cache template
		m.templateMutex.Lock()
		m.templates[name] = tmpl
		m.templateMutex.Unlock()
	}

	// Render template
	return tmpl.Execute(w, data)
}

// RenderWithBase renders a template with a base template
func (m *Manager) RenderWithBase(w io.Writer, baseName, contentName, contentPath string, data interface{}) error {
	// Get base template
	m.baseTemplateMutex.RLock()
	baseTemplate, baseExists := m.baseTemplates[baseName]
	m.baseTemplateMutex.RUnlock()

	if !baseExists {
		return fmt.Errorf("base template %s not found", baseName)
	}

	// Check if content template is cached
	m.templateMutex.RLock()
	contentTemplate, contentExists := m.templates[contentName]
	m.templateMutex.RUnlock()

	// Parse content template if not cached
	if !contentExists {
		var err error
		contentTemplate, err = template.ParseFiles(contentPath)
		if err != nil {
			return fmt.Errorf("failed to parse content template %s: %w", contentPath, err)
		}

		// Cache content template
		m.templateMutex.Lock()
		m.templates[contentName] = contentTemplate
		m.templateMutex.Unlock()
	}

	// Clone base template
	tmpl, err := baseTemplate.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone base template: %w", err)
	}

	// Add content template
	_, err = tmpl.AddParseTree("content", contentTemplate.Tree)
	if err != nil {
		return fmt.Errorf("failed to add content to base template: %w", err)
	}

	// Render combined template
	return tmpl.Execute(w, data)
}

// LoadTemplatesFromDir loads all templates from a directory
func (m *Manager) LoadTemplatesFromDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process template files
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			name := filepath.Base(path)
			name = name[:len(name)-len(filepath.Ext(name))] // Remove extension

			// Parse and store template
			if err := m.RegisterTemplate(name, path); err != nil {
				return err
			}
		}

		return nil
	})
}
