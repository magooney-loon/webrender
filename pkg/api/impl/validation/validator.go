package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	types "github.com/magooney-loon/webserver/types/api"
)

type validator struct{}

// New creates a new validator
func New() types.RequestDecoder {
	return &validator{}
}

// DecodeJSON decodes JSON request body
func (v *validator) DecodeJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// If the destination implements Validator, validate it
	if validator, ok := dst.(types.Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DecodeQuery decodes query parameters
func (v *validator) DecodeQuery(r *http.Request, dst interface{}) error {
	values := r.URL.Query()
	val := reflect.ValueOf(dst)

	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get("query")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		name := parts[0]
		required := len(parts) > 1 && parts[1] == "required"

		if !values.Has(name) {
			if required {
				return fmt.Errorf("missing required query parameter: %s", name)
			}
			continue
		}

		value := values.Get(name)
		if err := setField(field, value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", name, err)
		}
	}

	// If the destination implements Validator, validate it
	if validator, ok := dst.(types.Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// setField sets a field value from a string
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Slice:
		values := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(values), len(values))
		for i, v := range values {
			if err := setField(slice.Index(i), strings.TrimSpace(v)); err != nil {
				return err
			}
		}
		field.Set(slice)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
