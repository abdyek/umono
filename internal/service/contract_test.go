package service

import (
	"errors"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newContractTestService(t *testing.T) (*ContractService, map[string]models.Contract, map[string]models.SitePage) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Contract{}, &models.SitePage{}); err != nil {
		t.Fatalf("automigrate contracts and pages: %v", err)
	}

	contracts := []models.Contract{
		{Org: "umono", Name: "shop", Major: 1, Minor: 1, Source: "s", Checksum: "x"},
		{Org: "umono", Name: "shop", Major: 1, Minor: 2, Source: "s", Checksum: "x"},
		{Org: "umono", Name: "shop", Major: 1, Minor: 3, Source: "s", Checksum: "x"},
		{Org: "umono", Name: "shop", Major: 2, Minor: 0, Source: "s", Checksum: "x"},
		{Org: "umono", Name: "page", Major: 0, Minor: 1, Source: "s", Checksum: "x"},
		{Org: "umono", Name: "page", Major: 0, Minor: 2, Source: "s", Checksum: "x"},
	}

	stored := map[string]models.Contract{}
	for i := range contracts {
		c := contracts[i]
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create contract: %v", err)
		}
		stored[c.Ref()] = c
	}

	shopV12ID := stored["umono/shop@v1.2"].ID
	shopV13ID := stored["umono/shop@v1.3"].ID
	pageV01ID := stored["umono/page@v0.1"].ID

	layout := "page"
	pages := []models.SitePage{
		{Name: "A", Slug: "a", Enabled: true, ContractID: &shopV12ID, LayoutID: &layout},
		{Name: "B", Slug: "b", Enabled: false, ContractID: &shopV13ID, LayoutID: &layout},
		{Name: "C", Slug: "c", Enabled: true},
		{Name: "D", Slug: "d", Enabled: true, ContractID: &pageV01ID, LayoutID: &layout},
	}

	storedPages := map[string]models.SitePage{}
	for i := range pages {
		p := pages[i]
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create page: %v", err)
		}
		storedPages[p.Slug] = p
	}

	return NewContractService(repository.NewContractRepository(db)), stored, storedPages
}

func contractOf(org, name string, major, minor uint) models.Contract {
	return models.Contract{Org: org, Name: name, Major: major, Minor: minor}
}

func TestContractServiceBase(t *testing.T) {
	svc, _, _ := newContractTestService(t)

	cases := []struct {
		name   string
		g      models.ContractGroup
		wantOK bool
		want   uint
	}{
		{
			name:   "shop v1 has three contracts and bound pages",
			g:      contractOf("umono", "shop", 1, 0).Group(),
			wantOK: true,
			want:   3,
		},
		{
			name:   "shop v2 has no bound pages",
			g:      contractOf("umono", "shop", 2, 0).Group(),
			wantOK: false,
		},
		{
			name:   "page v0.1 has bound page",
			g:      contractOf("umono", "page", 0, 1).Group(),
			wantOK: true,
			want:   1,
		},
		{
			name:   "page v0.2 has no bound pages",
			g:      contractOf("umono", "page", 0, 2).Group(),
			wantOK: false,
		},
		{
			name:   "acme/none is missing",
			g:      contractOf("acme", "none", 1, 0).Group(),
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base, ok := svc.Base(tc.g)
			if ok != tc.wantOK {
				t.Fatalf("Base ok = %v, want %v", ok, tc.wantOK)
			}
			if tc.wantOK && base != tc.want {
				t.Fatalf("Base count = %d, want %d", base, tc.want)
			}
		})
	}
}

func TestContractServiceSelectable(t *testing.T) {
	svc, _, _ := newContractTestService(t)

	cases := []struct {
		name string
		c    models.Contract
		want bool
	}{
		{name: "shop v1.3 is the highest in v1", c: contractOf("umono", "shop", 1, 3), want: true},
		{name: "shop v1.2 has a higher version in v1", c: contractOf("umono", "shop", 1, 2), want: false},
		{name: "shop v1.4 has no higher version in v1", c: contractOf("umono", "shop", 1, 4), want: true},
		{name: "shop v2.0 is the only one in v2", c: contractOf("umono", "shop", 2, 0), want: true},
		{name: "page v0.1 is the only one in v0.1", c: contractOf("umono", "page", 0, 1), want: true},
		{name: "page v0.2 is the only one in v0.2", c: contractOf("umono", "page", 0, 2), want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := svc.Selectable(tc.c); got != tc.want {
				t.Fatalf("Selectable = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContractServiceOffered(t *testing.T) {
	svc, records, _ := newContractTestService(t)

	shopV1 := contractOf("umono", "shop", 1, 0).Group()
	pageV01 := contractOf("umono", "page", 0, 1).Group()
	pageV02 := contractOf("umono", "page", 0, 2).Group()
	acmeNone := contractOf("acme", "none", 1, 0).Group()

	activeShopV12 := []models.ThemeOption{
		{Path: "acme/shop/1.0.0", ContractID: "umono/shop@v1.2"},
	}
	activePageV01 := []models.ThemeOption{
		{Path: "acme/page/1.0.0", ContractID: "umono/page@v0.1"},
	}

	t.Run("shop v1 with nil active returns the highest contract", func(t *testing.T) {
		got, err := svc.Offered(shopV1, nil)
		if err != nil {
			t.Fatalf("Offered error: %v", err)
		}
		want := records["umono/shop@v1.3"]
		if got.ID != want.ID {
			t.Fatalf("Offered ID = %d, want %d", got.ID, want.ID)
		}
	})

	t.Run("shop v1 with active shop v1.2 returns shop v1.2", func(t *testing.T) {
		got, err := svc.Offered(shopV1, activeShopV12)
		if err != nil {
			t.Fatalf("Offered error: %v", err)
		}
		want := records["umono/shop@v1.2"]
		if got.ID != want.ID {
			t.Fatalf("Offered ID = %d, want %d", got.ID, want.ID)
		}
	})

	t.Run("shop v1 with active page returns the highest shop", func(t *testing.T) {
		got, err := svc.Offered(shopV1, activePageV01)
		if err != nil {
			t.Fatalf("Offered error: %v", err)
		}
		want := records["umono/shop@v1.3"]
		if got.ID != want.ID {
			t.Fatalf("Offered ID = %d, want %d", got.ID, want.ID)
		}
	})

	t.Run("page v0.2 with nil active returns page v0.2", func(t *testing.T) {
		got, err := svc.Offered(pageV02, nil)
		if err != nil {
			t.Fatalf("Offered error: %v", err)
		}
		want := records["umono/page@v0.2"]
		if got.ID != want.ID {
			t.Fatalf("Offered ID = %d, want %d", got.ID, want.ID)
		}
	})

	t.Run("page v0.1 with nil active returns page v0.1", func(t *testing.T) {
		got, err := svc.Offered(pageV01, nil)
		if err != nil {
			t.Fatalf("Offered error: %v", err)
		}
		want := records["umono/page@v0.1"]
		if got.ID != want.ID {
			t.Fatalf("Offered ID = %d, want %d", got.ID, want.ID)
		}
	})

	t.Run("acme/none with nil active returns ErrContractNotFound", func(t *testing.T) {
		_, err := svc.Offered(acmeNone, nil)
		if !errors.Is(err, ErrContractNotFound) {
			t.Fatalf("Offered error = %v, want ErrContractNotFound", err)
		}
	})
}

func TestContractServiceUncoveredPages(t *testing.T) {
	svc, _, _ := newContractTestService(t)

	cases := []struct {
		name string
		c    models.Contract
		want []string
	}{
		{name: "shop v1.2 covers a but not b", c: contractOf("umono", "shop", 1, 2), want: []string{"b"}},
		{name: "shop v1.3 covers everything", c: contractOf("umono", "shop", 1, 3), want: []string{}},
		{name: "shop v1.1 covers nothing in v1 group", c: contractOf("umono", "shop", 1, 1), want: []string{"a", "b"}},
		{name: "shop v2.0 has no bound pages", c: contractOf("umono", "shop", 2, 0), want: []string{}},
		{name: "page v0.1 covers its bound page", c: contractOf("umono", "page", 0, 1), want: []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.UncoveredPages(tc.c)
			slugs := make([]string, 0, len(got))
			for _, p := range got {
				slugs = append(slugs, p.Slug)
			}
			if len(slugs) != len(tc.want) {
				t.Fatalf("UncoveredPages len = %d, want %d (%v)", len(slugs), len(tc.want), slugs)
			}
			for i, slug := range tc.want {
				if slugs[i] != slug {
					t.Fatalf("UncoveredPages[%d] = %q, want %q", i, slugs[i], slug)
				}
			}
		})
	}
}
