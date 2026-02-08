package server

import (
	"fmt"
	"html/template"
	"reflect"
	"time"
)

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Int || rv.Kind() == reflect.Int64 {
			return int(rv.Int())
		}
		return 0
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("15:04:05")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02T15:04")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"localNumber": func(n float32) string {
			return fmt.Sprintf("%.0f", n)
		},
		"seq": func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i
			}
			return result
		},
		"now": func() time.Time {
			return time.Now()
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key, _ := values[i].(string)
				dict[key] = values[i+1]
			}
			return dict
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"gt": func(a, b interface{}) bool {
			return toInt(a) > toInt(b)
		},
		"gte": func(a, b interface{}) bool {
			return toInt(a) >= toInt(b)
		},
		"lt": func(a, b interface{}) bool {
			return toInt(a) < toInt(b)
		},
		"lte": func(a, b interface{}) bool {
			return toInt(a) <= toInt(b)
		},
		"le": func(a, b interface{}) bool {
			return toInt(a) <= toInt(b)
		},
		"ge": func(a, b interface{}) bool {
			return toInt(a) >= toInt(b)
		},
		"deref": func(p interface{}) interface{} {
			rv := reflect.ValueOf(p)
			if rv.Kind() == reflect.Ptr && !rv.IsNil() {
				return rv.Elem().Interface()
			}
			return p
		},
	}
}
