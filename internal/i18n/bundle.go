package i18n

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const MissingPrefix = "[[missing: "

type Locale struct {
	Name         string
	Lang         string
	Dir          string
	Translations map[string]string
}

type LocaleOption struct {
	Code string
	Name string
}

type Bundle struct {
	fallback string
	locales  map[string]*Locale
}

type Translator struct {
	bundle *Bundle
	locale *Locale
}

func LoadBundle(fsys fs.FS, fallback string) (*Bundle, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	bundle := &Bundle{
		fallback: fallback,
		locales:  make(map[string]*Locale),
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		file, err := fsys.Open(entry.Name())
		if err != nil {
			return nil, err
		}

		locale, err := parseLocale(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("parse locale %s: %w", entry.Name(), err)
		}
		if locale.Lang == "" {
			return nil, fmt.Errorf("parse locale %s: missing lang", entry.Name())
		}
		if locale.Name == "" {
			return nil, fmt.Errorf("parse locale %s: missing name", entry.Name())
		}
		if locale.Dir == "" {
			return nil, fmt.Errorf("parse locale %s: missing dir", entry.Name())
		}

		bundle.locales[locale.Lang] = locale
	}

	if _, ok := bundle.locales[fallback]; !ok {
		return nil, fmt.Errorf("fallback locale %q not found", fallback)
	}

	return bundle, nil
}

func (b *Bundle) Fallback() string {
	return b.fallback
}

func (b *Bundle) Translator(lang string) *Translator {
	return &Translator{
		bundle: b,
		locale: b.Locale(lang),
	}
}

func (b *Bundle) Locale(lang string) *Locale {
	if locale, ok := b.locales[lang]; ok {
		return locale
	}
	return b.locales[b.fallback]
}

func (b *Bundle) HasLocale(lang string) bool {
	_, ok := b.locales[lang]
	return ok
}

func (b *Bundle) SupportedLocales() []LocaleOption {
	options := make([]LocaleOption, 0, len(b.locales))
	codes := make([]string, 0, len(b.locales))
	for code := range b.locales {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		locale := b.locales[code]
		options = append(options, LocaleOption{
			Code: code,
			Name: locale.Name,
		})
	}

	return slices.Clone(options)
}

func (t *Translator) T(key string) string {
	if t == nil || t.bundle == nil || t.locale == nil {
		return Missing(key)
	}

	if value, ok := t.locale.Translations[key]; ok {
		return value
	}

	fallback := t.bundle.locales[t.bundle.fallback]
	if fallback != nil {
		if value, ok := fallback.Translations[key]; ok {
			return value
		}
	}

	return Missing(key)
}

func (t *Translator) Lang() string {
	if t == nil || t.locale == nil || t.locale.Lang == "" {
		return ""
	}
	return t.locale.Lang
}

func (t *Translator) Dir() string {
	if t == nil || t.locale == nil || t.locale.Dir == "" {
		return ""
	}
	return t.locale.Dir
}

func (t *Translator) Locale() *Locale {
	if t == nil {
		return nil
	}
	return t.locale
}

func Missing(key string) string {
	return MissingPrefix + key + "]]"
}

func parseLocale(r io.Reader) (*Locale, error) {
	locale := &Locale{
		Translations: make(map[string]string),
	}

	scanner := bufio.NewScanner(r)
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		key, value, err := parseAssignment(line)
		if err != nil {
			return nil, err
		}

		switch section {
		case "":
			switch key {
			case "name":
				locale.Name = value
			case "lang":
				locale.Lang = value
			case "dir":
				locale.Dir = value
			}
		case "translations":
			locale.Translations[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return locale, nil
}

func parseAssignment(line string) (string, string, error) {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return "", "", fmt.Errorf("invalid assignment: %q", line)
	}

	key := strings.TrimSpace(line[:idx])
	rawValue := strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", fmt.Errorf("missing key in assignment: %q", line)
	}

	if strings.HasPrefix(key, "\"") && strings.HasSuffix(key, "\"") {
		unquotedKey, err := strconv.Unquote(key)
		if err != nil {
			return "", "", fmt.Errorf("invalid quoted key %q: %w", key, err)
		}
		key = unquotedKey
	}

	value, err := strconv.Unquote(rawValue)
	if err != nil {
		return "", "", fmt.Errorf("invalid string value for %q: %w", key, err)
	}

	return key, value, nil
}
