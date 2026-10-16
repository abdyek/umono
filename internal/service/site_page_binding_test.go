package service

import (
	"errors"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func bindingOf(t *testing.T, errs []error) *BindingError {
	t.Helper()
	for _, err := range errs {
		var be *BindingError
		if errors.As(err, &be) {
			return be
		}
	}
	t.Fatalf("expected a *BindingError in %v, got none", errs)
	return nil
}

func TestCreatePageChecksTheBinding(t *testing.T) {
	sps, _, db := newRenderTestServices(t)
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	_, errs := sps.Create(models.SitePage{
		Name:       "A",
		Slug:       "a",
		Enabled:    true,
		Content:    "# Not a layout call",
		ContractID: &rec.ID,
		LayoutID:   strp("page"),
	})
	if e := bindingOf(t, errs); e.Kind != BindingRoot {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingRoot)
	}

	created, errs := sps.Create(models.SitePage{
		Name:       "B",
		Slug:       "b",
		Enabled:    true,
		Content:    fullPageCall,
		ContractID: &rec.ID,
		LayoutID:   strp("page"),
	})
	if len(errs) != 0 {
		t.Fatalf("len(errs) = %d, want 0; errs = %v", len(errs), errs)
	}
	if created.ContractID == nil {
		t.Fatalf("created.ContractID = nil, want %d", rec.ID)
	}
	if *created.ContractID != rec.ID {
		t.Fatalf("*created.ContractID = %d, want %d", *created.ContractID, rec.ID)
	}

	_, errs = sps.Create(models.SitePage{
		Name:    "C",
		Slug:    "c",
		Enabled: true,
		Content: "{{ UMONO_LAYOUT }}\n",
	})
	if e := bindingOf(t, errs); e.Kind != BindingUnboundCall {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingUnboundCall)
	}
}

func TestUpdatePageChecksTheBinding(t *testing.T) {
	sps, _, db := newRenderTestServices(t)
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	created, errs := sps.Create(models.SitePage{
		Name:       "B",
		Slug:       "b",
		Enabled:    true,
		Content:    fullPageCall,
		ContractID: &rec.ID,
		LayoutID:   strp("page"),
	})
	if len(errs) != 0 {
		t.Fatalf("len(errs) = %d, want 0; errs = %v", len(errs), errs)
	}

	created.Content = "{{ UMONO_LAYOUT site-name = \"S\" }}\n"
	_, errs = sps.Update(created)
	if e := bindingOf(t, errs); e.Kind != BindingMissingParameter {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingMissingParameter)
	}

	var got models.SitePage
	if err := db.First(&got, created.ID).Error; err != nil {
		t.Fatalf("db.First() error = %v", err)
	}
	if got.Content != fullPageCall {
		t.Fatalf("got.Content = %q, want %q", got.Content, fullPageCall)
	}
}
