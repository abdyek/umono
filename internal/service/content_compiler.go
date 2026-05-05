package service

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
)

var ErrContentCompilerNotConfigured = errors.New("content compiler is not configured")

type ContentCompiler struct {
	mu      sync.Mutex
	compono compono.Compono
}

func NewContentCompiler(components []models.Component) (*ContentCompiler, error) {
	c := &ContentCompiler{
		compono: compono.New(),
	}

	if err := c.LoadGlobalComponents(components); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *ContentCompiler) Compile(source string) (string, error) {
	return c.compile(source, c.convertContext())
}

func (c *ContentCompiler) CompileWithContext(source string, values map[string]any) (string, error) {
	return c.compile(source, c.convertContext(values))
}

func (c *ContentCompiler) PreviewComponent(name, source string) (string, error) {
	name = strings.TrimSpace(name)

	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	if err := c.compono.Convert([]byte("{{"+name+"}}"), &buf,
		compono.WithContext(c.convertContext()),
		compono.WithGlobalComponent(name, []byte(strings.TrimSpace(source))),
	); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *ContentCompiler) LoadGlobalComponents(components []models.Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, component := range components {
		if err := c.loadGlobalComponent(component); err != nil {
			return err
		}
	}

	return nil
}

func (c *ContentCompiler) LoadGlobalComponent(component models.Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.loadGlobalComponent(component)
}

func (c *ContentCompiler) RemoveGlobalComponent(component models.Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.removeGlobalComponent(component)
}

func (c *ContentCompiler) ReloadGlobalComponent(component models.Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.removeGlobalComponent(component); err != nil {
		return err
	}

	return c.loadGlobalComponent(component)
}

func (c *ContentCompiler) compile(source string, context map[string]any) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	if err := c.compono.Convert([]byte(source), &buf, compono.WithContext(context)); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *ContentCompiler) loadGlobalComponent(component models.Component) error {
	return c.compono.RegisterGlobalComponent(component.Name, []byte(component.Content))
}

func (c *ContentCompiler) removeGlobalComponent(component models.Component) error {
	return c.compono.UnregisterGlobalComponent(component.Name)
}

func (c *ContentCompiler) convertContext(values ...map[string]any) map[string]any {
	context := map[string]any{}

	for _, source := range values {
		for key, value := range source {
			context[key] = value
		}
	}

	return context
}
