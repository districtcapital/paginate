// Copyright District Capital Inc 2019
// All rights reserved.

package form

import (
	"reflect"
	"testing"

	"github.com/districtcapital/paginate"
	"github.com/stretchr/testify/assert"
)

func TestPopulate(t *testing.T) {
	type req struct {
		Page     int
		PageSize int
		ID       int16 `clause:"where"`
		Age      uint  `clause:"where"`
		Height   int64 `clause:"where"`
		Select   string
		OrderBy  []string
	}
	r := &req{
		Page:     7,
		PageSize: 25,
		ID:       99,
		Age:      37,
		Height:   550,
		Select:   "name",
		OrderBy:  []string{"id", "age"},
	}
	q := &paginate.Query{}
	ToQuery(r, q)
	assert.Equal(t, uint32(7), q.Page)
	assert.Equal(t, uint16(25), q.PageSize)
	if want, got := map[string]interface{}{"age": uint(37), "height": int64(550), "id": int16(99)}, q.WhereArgs; !reflect.DeepEqual(got, want) {
		t.Fatalf("Where maps do not match: want = %v got = %v", want, got)
	}
	assert.Equal(t, []string{"name"}, q.Select)
	assert.Equal(t, []string{"id", "age"}, q.OrderBy)
}

func TestPopulateAlternateForm(t *testing.T) {
	type req struct {
		Page     uint64
		PageSize uint32
		ZZZ      int16 `clause:"where, id"`
		XXX      *uint `clause:"where,age"`
		Height   int8  `clause:"where"`
		Select   []string
		OrderBy  string
	}
	i := uint(69)
	r := &req{
		Page:     79,
		PageSize: 25,
		ZZZ:      99,
		XXX:      &i,
		// Not including field 'Height' will also not include it in the final map.
		Select:  []string{"one", "two", "three"},
		OrderBy: "order-me",
	}
	q := &paginate.Query{}
	ToQuery(r, q)
	assert.Equal(t, uint32(79), q.Page)
	assert.Equal(t, uint16(25), q.PageSize)
	x := uint(69)
	if want, got := map[string]interface{}{"age": &x, "id": int16(99)}, q.WhereArgs; !reflect.DeepEqual(want, got) {
		t.Fatalf("Where maps do not match, want = %v got = %v", want, got)
	}
	assert.Equal(t, []string{"one", "two", "three"}, q.Select)
	assert.Equal(t, []string{"order-me"}, q.OrderBy)
}

func TestPopulateInvalid(t *testing.T) {
	type req struct {
		Page     string
		PageSize []byte
		OrderBy  map[string]string
		Selectme interface{} `clause:"select"`
	}
	r := &req{
		Page:     "10",
		PageSize: []byte("32"),
		OrderBy:  map[string]string{"age": "meh"},
		Selectme: "foo",
	}
	q := &paginate.Query{}
	ToQuery(r, q)
	// Nothing was populated.
	assert.Equal(t, &paginate.Query{}, q)
}

func TestPatchLikeQuery(t *testing.T) {
	c := paginate.Config{
		Where: map[string]string{"name": "like ?", "id": "= ?"},
	}
	q := paginate.Query{
		WhereArgs: map[string]interface{}{"name": "bob", "id": 38, "bogus": "blah"},
	}
	PatchLikeQuery(&c, &q)
	assert.Equal(t, 3, len(q.WhereArgs))          // No field was added or removed.
	assert.Equal(t, "%bob%", q.WhereArgs["name"]) // Name was patched.
	assert.Equal(t, "blah", q.WhereArgs["bogus"]) // Not patched (does not match).
	assert.Equal(t, 38, q.WhereArgs["id"])        // Not patched (not string).

	// Calling it again does not add extra "%"s.
	PatchLikeQuery(&c, &q)
	assert.Equal(t, 3, len(q.WhereArgs))          // No field was added or removed.
	assert.Equal(t, "%bob%", q.WhereArgs["name"]) // Name was patched.
	assert.Equal(t, "blah", q.WhereArgs["bogus"]) // Not patched (does not match).
	assert.Equal(t, 38, q.WhereArgs["id"])        // Not patched (not string).
}

func TestSnakeCase(t *testing.T) {
	assert.Equal(t, "", snakeCase(""))
	assert.Equal(t, "id", snakeCase("ID"))
	assert.Equal(t, "user_id", snakeCase("UserID"))
	assert.Equal(t, "user_id", snakeCase("UserId"))
	assert.Equal(t, "http", snakeCase("HTTP"))
	assert.Equal(t, "httpdirectory", snakeCase("HTTPDirectory"))
	assert.Equal(t, "httpdirectory_id", snakeCase("HTTPDirectoryID"))
	assert.Equal(t, "copy_url", snakeCase("CopyURL"))
	assert.Equal(t, "path_url", snakeCase("PathUrl"))
	assert.Equal(t, "funky_string___yo", snakeCase("Funky_String___yo"))
	// Either we make this one be nice ("copy_url_to_id") or we have to
	// special case things like "UserId" and "PathUrl".
	assert.Equal(t, "copy_urlto_id", snakeCase("CopyURLtoID"))
}
