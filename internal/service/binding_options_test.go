package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func paramNames(o BindingOption) []string {
	names := make([]string, len(o.Params))
	for i, p := range o.Params {
		names[i] = p.Name
	}
	return names
}

func TestBindingOptionsOfTheOfferedRecord(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	old := storeBlogContract(t, db)
	newer := storeContract(t, db, "example", "blog", 1, 1, string(contractSource(t)))

	opts, err := svc.BindingOptions(models.SitePage{})
	if err != nil {
		t.Fatalf("BindingOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1; opts = %v", len(opts), opts)
	}

	o := opts[0]
	if o.ContractID != newer.ID {
		t.Fatalf("o.ContractID = %d, want %d", o.ContractID, newer.ID)
	}
	if o.LayoutID != "post" {
		t.Fatalf("o.LayoutID = %q, want %q", o.LayoutID, "post")
	}
	wantValue := fmt.Sprintf("%d:post", newer.ID)
	if o.Value != wantValue {
		t.Fatalf("o.Value = %q, want %q", o.Value, wantValue)
	}
	if o.Group != "example/blog v1" {
		t.Fatalf("o.Group = %q, want %q", o.Group, "example/blog v1")
	}
	if o.Label != "example/blog v1.1 · post" {
		t.Fatalf("o.Label = %q, want %q", o.Label, "example/blog v1.1 · post")
	}
	if o.Old {
		t.Fatalf("o.Old = true, want false")
	}
	if !sameNames(paramNames(o), []string{"title", "body"}) {
		t.Fatalf("paramNames(o) = %v, want %v", paramNames(o), []string{"title", "body"})
	}
	if len(o.Params) == 0 {
		t.Fatalf("o.Params is empty")
	}
	if o.Params[0].Required {
		t.Fatalf("o.Params[0].Required = true, want false")
	}

	_ = old
}

func TestBindingOptionsKeepTheOldRecordOfThePage(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	old := storeBlogContract(t, db)
	storeContract(t, db, "example", "blog", 1, 1, string(contractSource(t)))

	page := bindPage(t, db, old.ID, "post")

	opts, err := svc.BindingOptions(page)
	if err != nil {
		t.Fatalf("BindingOptions: %v", err)
	}
	if len(opts) != 2 {
		t.Fatalf("len(opts) = %d, want 2; opts = %v", len(opts), opts)
	}

	var oldOpts []BindingOption
	for _, o := range opts {
		if o.Old {
			oldOpts = append(oldOpts, o)
		}
	}
	if len(oldOpts) != 1 {
		t.Fatalf("len(oldOpts) = %d, want 1; opts = %v", len(oldOpts), opts)
	}
	o := oldOpts[0]
	if o.ContractID != old.ID {
		t.Fatalf("o.ContractID = %d, want %d", o.ContractID, old.ID)
	}
	if o.LayoutID != "post" {
		t.Fatalf("o.LayoutID = %q, want %q", o.LayoutID, "post")
	}
	if o.Label != "example/blog v1.0 · post" {
		t.Fatalf("o.Label = %q, want %q", o.Label, "example/blog v1.0 · post")
	}
}

func TestBindingOptionsRaiseTheBase(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	storeBlogContract(t, db)
	newer := storeContract(t, db, "example", "blog", 1, 1, string(contractSource(t)))

	opts, err := svc.BindingOptions(models.SitePage{})
	if err != nil {
		t.Fatalf("BindingOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1; opts = %v", len(opts), opts)
	}

	o := opts[0]
	if !o.RaisesBase {
		t.Fatalf("o.RaisesBase = false, want true")
	}
	if o.NewBase != "v1.1" {
		t.Fatalf("o.NewBase = %q, want %q", o.NewBase, "v1.1")
	}
	if !sameNames(o.Unselectable, []string{"acme/minimal 0.1.0"}) {
		t.Fatalf("o.Unselectable = %v, want %v", o.Unselectable, []string{"acme/minimal 0.1.0"})
	}

	bindPage(t, db, newer.ID, "post")

	opts, err = svc.BindingOptions(models.SitePage{})
	if err != nil {
		t.Fatalf("BindingOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1; opts = %v", len(opts), opts)
	}
	if opts[0].RaisesBase {
		t.Fatalf("opts[0].RaisesBase = true, want false")
	}
}

func TestBindingOptionsWithAnActiveTheme(t *testing.T) {
	svc, _, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	c, err := svc.Activate(context.Background(), "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}

	opts, err := svc.BindingOptions(models.SitePage{})
	if err != nil {
		t.Fatalf("BindingOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1; opts = %v", len(opts), opts)
	}
	if opts[0].ContractID != c.Contract.ID {
		t.Fatalf("opts[0].ContractID = %d, want %d", opts[0].ContractID, c.Contract.ID)
	}
	if opts[0].RaisesBase {
		t.Fatalf("opts[0].RaisesBase = true, want false")
	}
}

func TestBindingOptionsWithoutRecords(t *testing.T) {
	svc, _, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	opts, err := svc.BindingOptions(models.SitePage{})
	if err != nil {
		t.Fatalf("BindingOptions: %v", err)
	}
	if len(opts) != 0 {
		t.Fatalf("len(opts) = %d, want 0; opts = %v", len(opts), opts)
	}
}
