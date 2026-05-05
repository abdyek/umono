package service

import "context"

type ContextProvider interface {
	BuildCompileContext(ctx context.Context, keys []string) (map[string]any, error)
}
