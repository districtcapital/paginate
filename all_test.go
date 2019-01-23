// Copyright District Capital Inc 2019
// All rights reserved.

package paginate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderBy(t *testing.T) {
	ob, err := orderBy(
		&Config{
			OrderableCols: []string{"id", "date"},
		},
		&Query{
			OrderBy: []string{"ID ASC", "dAte"},
		})
	assert.NoError(t, err)
	assert.Equal(t, "id asc, date", ob)

	ob, err = orderBy(
		&Config{
			OrderableCols: []string{"ID", "date"},
		},
		&Query{
			OrderBy: []string{" ID desc  ", ""},
		})
	assert.NoError(t, err)
	assert.Equal(t, "id desc", ob)

	// Invalid sort direction
	ob, err = orderBy(
		&Config{
			OrderableCols: []string{"ID", "date"},
		},
		&Query{
			OrderBy: []string{" ID goingup  "},
		})
	assert.Error(t, err)
	assert.Empty(t, ob)

	// Invalid field 'user_id'
	ob, err = orderBy(
		&Config{
			OrderableCols: []string{"ID", "date"},
		},
		&Query{
			OrderBy: []string{"user_id"},
		})
	assert.Error(t, err)
	assert.Empty(t, ob)

	// Too many fields.
	ob, err = orderBy(
		&Config{
			OrderableCols: []string{"ID", "date"},
		},
		&Query{
			OrderBy: []string{"id asc desc"},
		})
	assert.Error(t, err)
	assert.Empty(t, ob)

	// Cannot order by anything.
	ob, err = orderBy(
		&Config{},
		&Query{
			OrderBy: []string{"id"},
		})
	assert.Error(t, err)
	assert.Empty(t, ob)
}

func TestSelect(t *testing.T) {
	// Empty SelectableCols means "*"
	s, err := selectCols(
		&Config{},
		&Query{})
	assert.NoError(t, err)
	assert.Equal(t, "*", s)

	// If nothing is selected, everything *selectable* is selected.
	s, err = selectCols(
		&Config{
			SelectableCols: []string{"id", "date", "AGE"},
		},
		&Query{})
	assert.NoError(t, err)
	assert.Equal(t, "id, date, age", s)

	// Allowed subset is selected.
	s, err = selectCols(
		&Config{
			SelectableCols: []string{"id", "date", "AGE"},
		},
		&Query{
			Select: []string{"", "Date", "age", ""},
		})
	assert.NoError(t, err)
	assert.Equal(t, "date, age", s)

	// Any subset of "*" is allowed.
	s, err = selectCols(
		&Config{},
		&Query{
			Select: []string{"", " Date  ", "age", ""},
		})
	assert.NoError(t, err)
	assert.Equal(t, "date, age", s)

	// Cannot select is_admin.
	s, err = selectCols(
		&Config{
			SelectableCols: []string{"id", "date", "AGE"},
		},
		&Query{
			Select: []string{"Date", "age", "is_admin"},
		})
	assert.Error(t, err)
	assert.Equal(t, "", s)
}

func TestWhere(t *testing.T) {
	w, wa, err := where(&Config{
		Where: map[string]string{"id": "> ?", "age": "< ?"},
	},
		&Query{
			WhereArgs: map[string]interface{}{" ID ": 32, "aGe ": 69},
		})
	assert.NoError(t, err)
	// Maps are unsorted so we sort the where clause to ensure testable results.
	assert.Equal(t, "age < ? AND id > ?", w)
	assert.Equal(t, []interface{}{69, 32}, wa)

	// Age is not allowed.
	w, wa, err = where(&Config{
		Where: map[string]string{"id": "> ?", "gender": "= ?"},
	},
		&Query{
			WhereArgs: map[string]interface{}{"id": 32, "age": 69},
		})
	assert.Error(t, err)
	assert.Equal(t, "", w)
	assert.Equal(t, []interface{}(nil), wa)
}
