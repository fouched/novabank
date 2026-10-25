package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"sync"

	"github.com/fouched/novabank/notifications/platform/config"
)

type TemplateEngine struct {
	cfg   config.TemplatesConfig
	cache map[string]*template.Template
	mu    sync.RWMutex
}

func NewTemplateEngine(cfg config.TemplatesConfig) (*TemplateEngine, error) {
	engine := &TemplateEngine{
		cfg:   cfg,
		cache: make(map[string]*template.Template),
	}

	// Warm up the cache instantly on startup if enabled
	if cfg.CacheEnabled {
		if err := engine.compileAll(); err != nil {
			return nil, fmt.Errorf("failed to pre-compile layout templates: %w", err)
		}
	}

	return engine, nil
}

// Render compiles or fetches the requested layout file and populates it with dynamic data structures
func (e *TemplateEngine) Render(templateName string, data any) (string, error) {
	var t *template.Template
	var err error

	if e.cfg.CacheEnabled {
		e.mu.RLock()
		cachedTemplate, exists := e.cache[templateName]
		e.mu.RUnlock()

		if !exists {
			return "", fmt.Errorf("template %s not found in active memory cache", templateName)
		}
		t = cachedTemplate
	} else {
		// Hot reload straight from the filesystem folder layout path
		fullPath := filepath.Join(e.cfg.Directory, templateName)
		t, err = template.ParseFiles(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to parse template file dynamically from disk: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed executing html template data parsing transformations: %w", err)
	}

	return buf.String(), nil
}

func (e *TemplateEngine) compileAll() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Locate all HTML target files in your configurations directory tier boundary
	pattern := filepath.Join(e.cfg.Directory, "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := filepath.Base(file)
		t, err := template.ParseFiles(file)
		if err != nil {
			return fmt.Errorf("compile error on file %s: %w", file, err)
		}
		e.cache[name] = t
	}

	return nil
}
