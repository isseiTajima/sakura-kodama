package i18n

import (
	"embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"sync"
)

//go:embed locales/*.yaml
var localeFS embed.FS

type LocaleData map[string]interface{}

var (
	locales = make(map[string]LocaleData)
	mu      sync.RWMutex
)

// T returns the translated string for the given language and key.
func T(lang, key string) string {
	mu.RLock()
	data, ok := locales[lang]
	mu.RUnlock()

	if !ok {
		// Load on demand
		if err := loadLocale(lang); err != nil {
			return key
		}
		mu.RLock()
		data = locales[lang]
		mu.RUnlock()
	}

	val := getNested(data, key)
	if val == "" {
		// Fallback to ja if not found in requested lang
		if lang != "ja" {
			return T("ja", key)
		}
		return key
	}
	return val
}

// TVariant returns a random variant for the given key.
func TVariant(lang, key string) []string {
	mu.RLock()
	data, ok := locales[lang]
	mu.RUnlock()

	if !ok {
		if err := loadLocale(lang); err != nil {
			return []string{key}
		}
		mu.RLock()
		data = locales[lang]
		mu.RUnlock()
	}

	val := getNestedRaw(data, key)
	if val == nil {
		if lang != "ja" {
			return TVariant("ja", key)
		}
		return []string{key}
	}

	switch v := val.(type) {
	case []interface{}:
		res := make([]string, len(v))
		for i, s := range v {
			res[i] = fmt.Sprint(s)
		}
		return res
	case string:
		return []string{v}
	default:
		return []string{fmt.Sprint(v)}
	}
}

func loadLocale(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := locales[lang]; ok {
		return nil
	}

	filename := fmt.Sprintf("locales/%s.yaml", lang)
	data, err := localeFS.ReadFile(filename)
	if err != nil {
		return err
	}

	var localeData LocaleData
	if err := yaml.Unmarshal(data, &localeData); err != nil {
		return err
	}

	locales[lang] = localeData
	return nil
}

func getNested(data LocaleData, key string) string {
	val := getNestedRaw(data, key)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func getNestedRaw(data LocaleData, key string) interface{} {
	parts := strings.Split(key, ".")
	var current interface{} = data

	for _, part := range parts {
		if m, ok := current.(LocaleData); ok {
			current = m[part]
		} else if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}
