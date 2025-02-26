package template

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// CommonFuncs returns a map of commonly used template functions
func CommonFuncs() template.FuncMap {
	return template.FuncMap{
		// String helpers
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// Date/Time helpers
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"now": time.Now,
		"since": func(t time.Time) string {
			return time.Since(t).Round(time.Second).String()
		},
		"dateFormat": func(format string, t time.Time) string {
			return t.Format(format)
		},

		// Number helpers
		"inc": func(i int) int {
			return i + 1
		},
		"dec": func(i int) int {
			return i - 1
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Slice helpers
		"join": strings.Join,
		"slice": func(start, end int, items interface{}) interface{} {
			switch items := items.(type) {
			case []string:
				if end > len(items) {
					end = len(items)
				}
				return items[start:end]
			case []interface{}:
				if end > len(items) {
					end = len(items)
				}
				return items[start:end]
			default:
				return items
			}
		},

		// Conditional helpers
		"ternary": func(condition bool, trueVal, falseVal interface{}) interface{} {
			if condition {
				return trueVal
			}
			return falseVal
		},
		"default": func(value, defaultValue interface{}) interface{} {
			if value == nil {
				return defaultValue
			}
			return value
		},

		// URL helpers
		"urlEncode": template.URLQueryEscaper,
		"pathJoin": func(elem ...string) string {
			return strings.Join(elem, "/")
		},

		// HTML helpers
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},

		// Debug helpers
		"dump": func(v interface{}) string {
			return fmt.Sprintf("%+v", v)
		},

		// Map helpers
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires pairs of key/value, got %d values", len(values))
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings, got %T", values[i])
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}
