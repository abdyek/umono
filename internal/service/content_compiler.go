package service

import (
	"bytes"
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/umono/internal/models"
)

var ErrContentCompilerNotConfigured = errors.New("content compiler is not configured")

type ContentCompiler struct {
	mu               sync.Mutex
	compono          compono.Compono
	contextProviders []ContextProvider
	globalComponents map[string]string
}

func NewContentCompiler(components []models.Component, contextProviders ...ContextProvider) (*ContentCompiler, error) {
	c := &ContentCompiler{
		compono:          compono.New(),
		contextProviders: contextProviders,
		globalComponents: map[string]string{},
	}

	if err := c.LoadGlobalComponents(components); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *ContentCompiler) Compile(source string) (string, error) {
	return c.CompileWithProviderContext(context.Background(), source)
}

func (c *ContentCompiler) CompileWithProviderContext(ctx context.Context, source string) (string, error) {
	if len(c.contextProviders) == 0 {
		return c.compile(source, c.convertContext())
	}

	contextValues, err := c.buildCompileContext(ctx, c.contextKeys(source))
	if err != nil {
		return "", err
	}

	return c.compile(source, contextValues)
}

func (c *ContentCompiler) CompileWithContext(source string, values map[string]any) (string, error) {
	return c.compile(source, c.convertContext(values))
}

func (c *ContentCompiler) PreviewComponent(name, source string) (string, error) {
	return c.PreviewComponentWithProviderContext(context.Background(), name, source)
}

func (c *ContentCompiler) PreviewComponentWithProviderContext(ctx context.Context, name, source string) (string, error) {
	name = strings.TrimSpace(name)

	if len(c.contextProviders) == 0 {
		return c.previewComponent(name, source, c.convertContext())
	}

	contextValues, err := c.buildCompileContext(ctx, c.contextKeys(source))
	if err != nil {
		return "", err
	}

	return c.previewComponent(name, source, contextValues)
}

func (c *ContentCompiler) previewComponent(name, source string, contextValues map[string]any) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	if err := c.compono.Convert([]byte("{{"+name+"}}"), &buf,
		compono.WithContext(contextValues),
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
	if err := c.compono.RegisterGlobalComponent(component.Name, []byte(component.Content)); err != nil {
		return err
	}
	c.globalComponents[component.Name] = component.Content
	return nil
}

func (c *ContentCompiler) removeGlobalComponent(component models.Component) error {
	if err := c.compono.UnregisterGlobalComponent(component.Name); err != nil {
		return err
	}
	delete(c.globalComponents, component.Name)
	return nil
}

func (c *ContentCompiler) buildCompileContext(ctx context.Context, keys []string) (map[string]any, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	values := map[string]any{}
	for _, provider := range c.contextProviders {
		if provider == nil {
			continue
		}

		provided, err := provider.BuildCompileContext(ctx, keys)
		if err != nil {
			return nil, err
		}

		for key, value := range provided {
			values[key] = value
		}
	}

	return values, nil
}

func (c *ContentCompiler) contextKeys(source string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	seen := map[string]struct{}{}
	c.collectContextKeys(source, ast.DefaultRootNode(), seen)
	sourceAsGlobal := ast.DefaultEmptyNode()
	sourceAsGlobal.SetRule(rule.NewGlobalCompDef())
	c.collectContextKeys(source, sourceAsGlobal, seen)

	for _, globalSource := range c.globalComponents {
		node := ast.DefaultEmptyNode()
		node.SetRule(rule.NewGlobalCompDef())
		c.collectContextKeys(globalSource, node, seen)
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *ContentCompiler) collectContextKeys(source string, root ast.Node, seen map[string]struct{}) {
	if strings.TrimSpace(source) == "" {
		return
	}

	parsed := c.compono.Parser().Parse([]byte(source), root)
	for _, node := range ast.FilterNodesInTree(parsed, isContextValueNode) {
		key := ast.GetContextKey(node)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
}

func isContextValueNode(node ast.Node) bool {
	return ast.IsRuleNameOneOf(node, []string{
		"context-ref",
		"comp-context-param",
		"comp-call-context-arg",
	})
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
