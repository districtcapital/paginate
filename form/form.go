// Copyright District Capital Inc 2019
// All rights reserved.

// Package form helps build a paginate.Query from a web request.
package form

import (
	"bytes"
	"reflect"
	"strings"
	"unicode"

	"github.com/districtcapital/paginate"
)

// ToQuery populates fields in paginate.Query if they are present
// in req. It looks for fields named PageSize, Page, Select, OrderBy, and Search
// (exactly) first and copies the values found without interpretation, but
// converting them to their appropriate types if they fit. Then it looks for
// annotations "clause" which can be "where" followed or not by an alternate
// field name (if none is given the snake case of the field is used) and
// populates the WhereArgs map. If it doesn't find what it's looking for, it
// does not change Query at all. See tests for examples.
func ToQuery(req interface{}, q *paginate.Query) {
	typ := reflect.TypeOf(req).Elem()
	val := reflect.ValueOf(req).Elem()

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)

		switch typeField.Name {
		case "Page":
			q.Page = uint32(getUint64(structField))
			continue
		case "PageSize":
			q.PageSize = uint16(getUint64(structField))
			continue
		case "OrderBy":
			q.OrderBy = appendString(q.OrderBy, structField)
			continue
		case "Select":
			q.Select = appendString(q.Select, structField)
			continue
		case "Search":
			q.Search = getString(structField)
		}
		// Fall through for other fields not recognized by name.
		// Now we look for `clause:"where"` or variants such as
		// `clause:"where,age"` and `clause:"where,id"` etc.

		fullClause := strings.Split(strings.ToLower(strings.TrimSpace(typeField.Tag.Get("clause"))), ",")
		clauseType := fullClause[0]
		var argName string
		if len(fullClause) == 2 {
			argName = strings.TrimSpace(fullClause[1])
		} else {
			argName = snakeCase(typeField.Name)
		}
		switch clauseType {
		case "where":
			if q.WhereArgs == nil {
				q.WhereArgs = make(map[string]interface{})
			}
			if structField.Interface() != reflect.Zero(structField.Type()).Interface() {
				q.WhereArgs[argName] = structField.Interface()
			}
		}
	}
}

// PatchLikeQuery changes the WhereArgs in the query if their matching
// Where fields contain the SQL keyword "LIKE" (or "like") so that the arguments
// are surrounded by "%". Matching fields are only changed if the WhereArgs are
// strings. Search is changed if it's not empty. If no matching like fields are
// present, the query is unmodified. PatchLikeQuery never changes Config.
// DEPRECATED: use paginate.PatchLikeQuery() instead.
func PatchLikeQuery(c *paginate.Config, q *paginate.Query) {
	paginate.PatchLikeQuery(c, q, true, true)
}

func getUint64(v reflect.Value) uint64 {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uint64(v.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(v.Int())
	}
	return 0
}

func getString(v reflect.Value) string {
	k := v.Kind()
	if k == reflect.String {
		return v.String()
	}
	return ""
}

// appendString appends to slice all strings from v and returns it.
func appendString(slice []string, v reflect.Value) []string {
	k := v.Kind()
	if k == reflect.String {
		return append(slice, v.String())
	}
	if (k == reflect.Slice || k == reflect.Array) && v.Type().Elem().Kind() == reflect.String {
		for i := 0; i < v.Len(); i++ {
			slice = append(slice, v.Index(i).String())
		}
	}
	return slice
}

// snakeCase transforms a CamelCase string into snake_case.
func snakeCase(s string) string {
	var buf bytes.Buffer

	if len(s) == 0 {
		return ""
	}
	var last rune
	w := func(r rune) {
		last = r
		buf.WriteRune(unicode.ToLower(r))
	}
	w(rune(s[0]))
	for i := 1; i < len(s); i++ {
		r := rune(s[i])
		if r == '_' {
			w(r)
			continue
		}
		if unicode.IsUpper(r) {
			if !unicode.IsUpper(last) && last != rune('_') {
				w(rune('_'))
				w(rune(s[i]))
				continue
			}
			w(rune(s[i]))
			continue
		}
		w(rune(s[i]))
	}
	return buf.String()
}
