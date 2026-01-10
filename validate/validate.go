package validate

import (
	"encoding"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator.New()}
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.applyDefaults(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "应用默认值失败: "+err.Error())
	}

	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (cv *CustomValidator) applyDefaults(i any) error {
	if i == nil {
		return nil
	}
	return cv.applyDefaultsRecursive(reflect.ValueOf(i))
}

func (cv *CustomValidator) applyDefaultsRecursive(v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}

// 在 applyDefaultsRecursive 的 switch 中补充分支
switch field.Kind() {
case reflect.Struct, reflect.Ptr:
    if err := cv.applyDefaultsRecursive(field); err != nil {
        return err
    }
case reflect.Slice, reflect.Array:
    for j := 0; j < field.Len(); j++ {
        if err := cv.applyDefaultsRecursive(field.Index(j)); err != nil {
            return err
        }
    }
case reflect.Map:
    // 仅处理 value
    for _, k := range field.MapKeys() {
        mv := field.MapIndex(k)
        // MapIndex 返回不可设置的 Value，需要拷贝后递归并写回
        vv := reflect.New(mv.Type()).Elem()
        vv.Set(mv)
        if err := cv.applyDefaultsRecursive(vv); err != nil {
            return err
        }
        field.SetMapIndex(k, vv)
    }
}
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		if err := cv.ensureFieldInitialized(field, fieldType); err != nil {
			return err
		}

		if defVal := fieldType.Tag.Get("default"); defVal != "" && cv.isZeroValue(field) {
			if err := cv.setFieldValue(field, defVal); err != nil {
				return fmt.Errorf("field %s: %w", fieldType.Name, err)
			}
		}

		switch field.Kind() {
		case reflect.Struct, reflect.Ptr:
			if err := cv.applyDefaultsRecursive(field); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cv *CustomValidator) ensureFieldInitialized(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.Ptr || !field.IsNil() {
		return nil
	}

	elemType := field.Type().Elem()
	if elemType.Kind() == reflect.Struct {
		if fieldType.Tag.Get("default") != "" || hasNestedDefault(elemType, nil) {
			field.Set(reflect.New(elemType))
		}
	}

	return nil
}

func (cv *CustomValidator) isZeroValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	case reflect.Bool:
		return !field.Bool()
	case reflect.Slice, reflect.Map, reflect.Interface, reflect.Ptr:
		return field.IsNil()
	default:
		return field.IsZero()
	}
}

func (cv *CustomValidator) setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	if field.Kind() != reflect.Interface && field.Kind() != reflect.Ptr {
		if ok, err := setViaTextUnmarshaler(field, value); ok {
			return err
		}
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type().PkgPath() == "time" && field.Type().Name() == "Duration" {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
			return nil
		}
		bitSize := field.Type().Bits()
		if bitSize == 0 {
			bitSize = 64
		}
		val, err := strconv.ParseInt(value, 10, bitSize)
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bitSize := field.Type().Bits()
		if bitSize == 0 {
			bitSize = 64
		}
		val, err := strconv.ParseUint(value, 10, bitSize)
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Float32, reflect.Float64:
		bitSize := field.Type().Bits()
		if bitSize == 0 {
			bitSize = 64
		}
		val, err := strconv.ParseFloat(value, bitSize)
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
	case reflect.Slice, reflect.Map, reflect.Array:
		return setFromJSON(field, value)
	case reflect.Struct:
		return setFromJSON(field, value)
	case reflect.Interface:
		var v any
		if err := json.Unmarshal([]byte(value), &v); err != nil {
			field.Set(reflect.ValueOf(value))
			return nil
		}
		field.Set(reflect.ValueOf(v))
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return cv.setFieldValue(field.Elem(), value)
	default:
		return fmt.Errorf("unsupported kind %s for default tag", field.Kind())
	}

	return nil
}

func setFromJSON(field reflect.Value, raw string) error {
	if raw == "" {
		return nil
	}

	target := reflect.New(field.Type())
	if err := json.Unmarshal([]byte(raw), target.Interface()); err != nil {
		return err
	}
	field.Set(target.Elem())
	return nil
}

func setViaTextUnmarshaler(field reflect.Value, value string) (bool, error) {
	if field.CanAddr() {
		if unmarshaler, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return true, unmarshaler.UnmarshalText([]byte(value))
		}
	}
	return false, nil
}

func hasNestedDefault(t reflect.Type, visited map[reflect.Type]struct{}) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	if visited == nil {
		visited = make(map[reflect.Type]struct{})
	}

	if _, ok := visited[t]; ok {
		return false
	}
	visited[t] = struct{}{}

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.Tag.Get("default") != "" {
			return true
		}

		switch ft.Type.Kind() {
		case reflect.Struct:
			if hasNestedDefault(ft.Type, visited) {
				return true
			}
		case reflect.Ptr:
			if ft.Type.Elem().Kind() == reflect.Struct && hasNestedDefault(ft.Type.Elem(), visited) {
				return true
			}
		}
	}

	return false
}

func (cv *CustomValidator) ValidateWithDefaults(i any) error {
	return cv.Validate(i)
}

func (cv *CustomValidator) SetDefault(i any, fieldName, defaultValue string) error {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return echo.NewHTTPError(http.StatusBadRequest, "参数无效")
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return echo.NewHTTPError(http.StatusBadRequest, "字段不存在: "+fieldName)
	}

	if !field.CanSet() {
		return echo.NewHTTPError(http.StatusBadRequest, "字段不可设置: "+fieldName)
	}

	if field.Kind() == reflect.Ptr && field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}

	if err := cv.setFieldValue(field, defaultValue); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "设置默认值失败: "+err.Error())
	}
	return nil
}
