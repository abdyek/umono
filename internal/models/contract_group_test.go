package models

import "testing"

func ct(org, name string, major, minor uint) Contract {
	return Contract{Org: org, Name: name, Major: major, Minor: minor}
}

func TestContractGroup(t *testing.T) {
	cases := []struct {
		name    string
		c       Contract
		want    ContractGroup
		wantStr string
	}{
		{
			name:    "v1.3 collapses to v1",
			c:       ct("umono", "e-commerce", 1, 3),
			want:    ContractGroup{Org: "umono", Name: "e-commerce", Major: 1, Minor: 0},
			wantStr: "umono/e-commerce v1",
		},
		{
			name:    "v1.0 stays v1",
			c:       ct("umono", "e-commerce", 1, 0),
			want:    ContractGroup{Org: "umono", Name: "e-commerce", Major: 1, Minor: 0},
			wantStr: "umono/e-commerce v1",
		},
		{
			name:    "v0.1 keeps minor",
			c:       ct("umono", "page", 0, 1),
			want:    ContractGroup{Org: "umono", Name: "page", Major: 0, Minor: 1},
			wantStr: "umono/page v0.1",
		},
		{
			name:    "v0.2 keeps minor",
			c:       ct("umono", "page", 0, 2),
			want:    ContractGroup{Org: "umono", Name: "page", Major: 0, Minor: 2},
			wantStr: "umono/page v0.2",
		},
		{
			name:    "v12.7 collapses to v12",
			c:       ct("acme", "blog", 12, 7),
			want:    ContractGroup{Org: "acme", Name: "blog", Major: 12, Minor: 0},
			wantStr: "acme/blog v12",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.c.Group()
			if got != tc.want {
				t.Errorf("Group() = %#v, want %#v", got, tc.want)
			}
			if got.String() != tc.wantStr {
				t.Errorf("Group().String() = %q, want %q", got.String(), tc.wantStr)
			}
		})
	}
}

func TestContractCovers(t *testing.T) {
	cases := []struct {
		name string
		a, b Contract
		want bool
	}{
		{
			name: "v1.3 covers v1.2",
			a:    ct("umono", "shop", 1, 3),
			b:    ct("umono", "shop", 1, 2),
			want: true,
		},
		{
			name: "v1.2 does not cover v1.3",
			a:    ct("umono", "shop", 1, 2),
			b:    ct("umono", "shop", 1, 3),
			want: false,
		},
		{
			name: "v1.2 covers v1.2",
			a:    ct("umono", "shop", 1, 2),
			b:    ct("umono", "shop", 1, 2),
			want: true,
		},
		{
			name: "v1.10 covers v1.9",
			a:    ct("umono", "shop", 1, 10),
			b:    ct("umono", "shop", 1, 9),
			want: true,
		},
		{
			name: "v2.0 does not cover v1.0",
			a:    ct("umono", "shop", 2, 0),
			b:    ct("umono", "shop", 1, 0),
			want: false,
		},
		{
			name: "v0.2 does not cover v0.1",
			a:    ct("umono", "page", 0, 2),
			b:    ct("umono", "page", 0, 1),
			want: false,
		},
		{
			name: "v0.1 does not cover v0.2",
			a:    ct("umono", "page", 0, 1),
			b:    ct("umono", "page", 0, 2),
			want: false,
		},
		{
			name: "v0.1 covers v0.1",
			a:    ct("umono", "page", 0, 1),
			b:    ct("umono", "page", 0, 1),
			want: true,
		},
		{
			name: "other org",
			a:    ct("acme", "shop", 1, 3),
			b:    ct("umono", "shop", 1, 2),
			want: false,
		},
		{
			name: "other name",
			a:    ct("umono", "store", 1, 3),
			b:    ct("umono", "shop", 1, 2),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Covers(tc.b); got != tc.want {
				t.Errorf("(%#v).Covers(%#v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestContractRef(t *testing.T) {
	cases := []struct {
		name string
		c    Contract
		want string
	}{
		{
			name: "major with minor collapses to major",
			c:    ct("umono", "e-commerce", 1, 3),
			want: "umono/e-commerce@v1.3",
		},
		{
			name: "major zero keeps minor",
			c:    ct("umono", "page", 0, 1),
			want: "umono/page@v0.1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.Ref(); got != tc.want {
				t.Errorf("Ref() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseContractRef(t *testing.T) {
	cases := []struct {
		s    string
		want Contract
	}{
		{s: "umono/e-commerce@v1.3", want: ct("umono", "e-commerce", 1, 3)},
		{s: "umono/page@v0.1", want: ct("umono", "page", 0, 1)},
		{s: "a1/b-2-c@v10.20", want: ct("a1", "b-2-c", 10, 20)},
	}

	for _, tc := range cases {
		t.Run(tc.s, func(t *testing.T) {
			got, err := ParseContractRef(tc.s)
			if err != nil {
				t.Errorf("ParseContractRef(%q) returned error: %v", tc.s, err)
			}
			if got != tc.want {
				t.Errorf("ParseContractRef(%q) = %#v, want %#v", tc.s, got, tc.want)
			}
			if got.Ref() != tc.s {
				t.Errorf("ParseContractRef(%q).Ref() = %q, want %q", tc.s, got.Ref(), tc.s)
			}
		})
	}
}

func TestParseContractRefRejects(t *testing.T) {
	cases := []string{
		"",
		"umono/page",
		"umono/page@0.1",
		"umono/page@v1",
		"umono/page@v1.2.3",
		"umono@v1.2",
		"Umono/page@v1.2",
		"umono/Page@v1.2",
		"umono/page@v01.2",
		"umono/page@v1.02",
		"umono/page@v-1.2",
		"umono/page @v1.2",
		"umono/page@v1.2 ",
		"-umono/page@v1.2",
		"umono/page-@v1.2",
		"umono//page@v1.2",
		"umono/pa--ge@v1.2",
		"umono/page@v99999999999999999999.0",
	}

	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			_, err := ParseContractRef(s)
			if err == nil {
				t.Errorf("ParseContractRef(%q) = nil error, want non-nil", s)
			}
		})
	}
}
