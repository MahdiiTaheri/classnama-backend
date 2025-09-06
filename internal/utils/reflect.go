package utils

import "reflect"

// ApplyPatch copies non-nil pointer fields from src to dst struct.
// skipFields can be used to exclude certain fields like "Version".
func ApplyPatch(dst, src any, skipFields ...string) {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Pointer || dstVal.IsNil() {
		return
	}
	dstVal = dstVal.Elem()

	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() != reflect.Struct {
		return
	}

	dstType := dstVal.Type()
	skip := map[string]struct{}{}
	for _, f := range skipFields {
		skip[f] = struct{}{}
	}

	// Precompute dst fields map: name -> index
	dstFieldIndex := make(map[string]int, dstVal.NumField())
	for i := 0; i < dstVal.NumField(); i++ {
		dstFieldIndex[dstType.Field(i).Name] = i
	}

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		fieldName := srcVal.Type().Field(i).Name

		if _, skipField := skip[fieldName]; skipField {
			continue
		}

		if srcField.Kind() == reflect.Pointer && !srcField.IsNil() {
			if dstIdx, ok := dstFieldIndex[fieldName]; ok {
				dstField := dstVal.Field(dstIdx)
				if dstField.CanSet() {
					dstField.Set(srcField.Elem())
				}
			}
		}
	}
}
