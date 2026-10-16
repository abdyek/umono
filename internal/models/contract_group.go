package models

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// ContractGroup is the group of a contract. Themes and pages are grouped by
// it and a group has at most one active theme. For major 0 every minor is a
// group of its own, so Minor is part of the group only when Major is 0.
type ContractGroup struct {
	Org   string
	Name  string
	Major uint
	Minor uint
}

// String returns the group key: "org/id vMAJOR", or "org/id v0.MINOR" for
// major 0.
func (g ContractGroup) String() string {
	if g.Major == 0 {
		return fmt.Sprintf("%s/%s v0.%d", g.Org, g.Name, g.Minor)
	}
	return fmt.Sprintf("%s/%s v%d", g.Org, g.Name, g.Major)
}

// Group returns the group of the contract.
func (c Contract) Group() ContractGroup {
	g := ContractGroup{Org: c.Org, Name: c.Name, Major: c.Major}
	if c.Major == 0 {
		g.Minor = c.Minor
	}
	return g
}

// Covers reports whether the contract covers other: both are in the same
// group and c's minor is not lower. A v0 contract covers only itself.
func (c Contract) Covers(other Contract) bool {
	return c.Group() == other.Group() && c.Minor >= other.Minor
}

// Ref returns the reference of the contract: org/id@vMAJOR.MINOR.
func (c Contract) Ref() string {
	return fmt.Sprintf("%s/%s@v%d.%d", c.Org, c.Name, c.Major, c.Minor)
}

// ErrInvalidContractRef is returned for a string that is not a contract reference.
var ErrInvalidContractRef = errors.New("invalid contract reference")

// contractRefPattern is the grammar of a contract reference, the same as themecore's.
var contractRefPattern = regexp.MustCompile(`^([a-z0-9]+(?:-[a-z0-9]+)*)/([a-z0-9]+(?:-[a-z0-9]+)*)@v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// ParseContractRef parses org/id@vMAJOR.MINOR. Only the identity fields of the
// returned contract are set.
func ParseContractRef(s string) (Contract, error) {
	m := contractRefPattern.FindStringSubmatch(s)
	if m == nil {
		return Contract{}, ErrInvalidContractRef
	}
	major, err := strconv.ParseUint(m[3], 10, 32)
	if err != nil {
		return Contract{}, ErrInvalidContractRef
	}
	minor, err := strconv.ParseUint(m[4], 10, 32)
	if err != nil {
		return Contract{}, ErrInvalidContractRef
	}
	return Contract{Org: m[1], Name: m[2], Major: uint(major), Minor: uint(minor)}, nil
}
