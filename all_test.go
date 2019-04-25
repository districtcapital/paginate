// Copyright District Capital Inc 2019
// All rights reserved.

package paginate

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestOrderByClause(t *testing.T) {
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

func TestSelectClause(t *testing.T) {
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

func TestWhereClause(t *testing.T) {
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

	// No where clause is okay too.
	w, wa, err = where(&Config{}, &Query{})
	assert.NoError(t, err)
	assert.Equal(t, "", w)
	assert.Equal(t, []interface{}(nil), wa)

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

func TestWhereWithSearch(t *testing.T) {
	c := &Config{
		Where: map[string]string{
			"first_name": "like ?",
			"last_name":  "like ?",
			"age":        "> ?",
			"status":     "= ?",
		},
	}
	w, wa, err := where(c,
		&Query{
			WhereArgs: map[string]interface{}{
				"age": 30,
			},
			Search: "augustus",
		})
	assert.NoError(t, err)
	assert.Equal(t, "age > ? AND (first_name like ? OR last_name like ?)", w)
	assert.Equal(t, []interface{}{30, "augustus", "augustus"}, wa)

	// Don't set any AND param.
	w, wa, err = where(c,
		&Query{
			Search: "augustus",
		})
	assert.NoError(t, err)
	assert.Equal(t, "first_name like ? OR last_name like ?", w)
	assert.Equal(t, []interface{}{"augustus", "augustus"}, wa)

	// Pin a LIKE element down.
	w, wa, err = where(c,
		&Query{
			WhereArgs: map[string]interface{}{
				"age":        22,
				"first_name": "Bob",
			},
			Search: "augustus",
		})
	assert.NoError(t, err)
	assert.Equal(t, "age > ? AND first_name like ? AND (first_name like ? OR last_name like ?)", w)
	assert.Equal(t, []interface{}{22, "Bob", "augustus", "augustus"}, wa)

	// Disallows search term
	c.DisallowSearchTerm = true
	_, _, err = where(c, &Query{Search: "augustus"})
	assert.Error(t, err)
	assert.Equal(t, "search term is disallowed by config", err.Error())
}

func TestSimple(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 3,
	}
	q := Query{
		Page: 1,
	}

	var results []dbModel
	res, err := Do(db, c, q, &results)
	assert.NoError(t, err)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(3), res.RowsAffected)
	subSlice := testData[:3]
	assert.Equal(t, subSlice, results)

	q.Page = 2
	res, err = Do(db, c, q, &results)
	assert.NoError(t, err)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(3), res.RowsAffected)
	subSlice = testData[3:6]
	assert.Equal(t, subSlice, results)

	q.Page = 3
	res, err = Do(db, c, q, &results)
	assert.NoError(t, err)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(1), res.RowsAffected)
	subSlice = testData[6:7]
	assert.Equal(t, subSlice, results)
}

func TestWhere(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 2,
		Where:           map[string]string{"age": "> ?"},
	}
	q := Query{
		Page:      1,
		WhereArgs: map[string]interface{}{"age": "3"},
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
			{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
		},
		{
			{ID: 3, Name: "Test Dude", Age: 7, IQ: 200},
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
		},
		{
			{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
			{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
		},
	})
}

func TestOrderBy(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		OrderableCols: []string{"age", "iq"},
	}
	q := Query{
		PageSize: 4,
		Page:     1,
		OrderBy:  []string{"age asc", " iq DESC "},
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 5, Name: "Blah", Age: 3, IQ: 100},
			{ID: 3, Name: "Test Dude", Age: 7, IQ: 200},
			{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
			{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
		},
		{
			{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
			{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
		},
	})
}

func TestWhereAndOrderBy(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 2,
		Where:           map[string]string{"age": "> ?"},
		OrderableCols:   []string{"iq"},
	}
	q := Query{
		Page:      1,
		WhereArgs: map[string]interface{}{"age": 15},
		OrderBy:   []string{"iq desc"},
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
			{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
		},
		{
			{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
			{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
		},
		{
			{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
		},
	})
}

func TestSmallPage(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{}
	q := Query{
		PageSize: 1,
		Page:     1,
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
		},
		{
			{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
		},
		{
			{ID: 3, Name: "Test Dude", Age: 7, IQ: 200},
		},
		{
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
		},
		{
			{ID: 5, Name: "Blah", Age: 3, IQ: 100},
		},
		{
			{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
		},
		{
			{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
		},
	})
}

func TestBigPage(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 100,
	}
	q := Query{
		Page: 1,
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
			{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
			{ID: 3, Name: "Test Dude", Age: 7, IQ: 200},
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
			{ID: 5, Name: "Blah", Age: 3, IQ: 100},
			{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
			{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
		},
	})
}

func TestNoResults(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		MaxPageSize: 100,
		Where:       map[string]string{"age": "> ?"},
	}
	q := Query{
		PageSize:  1000,
		Page:      1,
		WhereArgs: map[string]interface{}{"age": 99},
	}

	testPagination(t, db, c, q, nil)
}

func TestDefaultPageSize(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{}
	q := Query{
		Page: 1,
	}
	testPagination(t, db, c, q, [][]dbModel{testData})
}

func TestHugePageSize(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 1<<16 - 1,
		MaxPageSize:     1<<16 - 1,
	}
	q := Query{
		Page: 1,
	}
	testPagination(t, db, c, q, [][]dbModel{testData})
}

func TestSelect(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		SelectableCols: []string{"age", "name"},
	}
	q := Query{
		PageSize: 10,
		Page:     1,
	}
	testPagination(t, db, c, q, [][]dbModel{
		{
			{Name: "Don Jr", Age: 46},
			{Name: "Potranka", Age: 44},
			{Name: "Test Dude", Age: 7},
			{Name: "Meh", Age: 77},
			{Name: "Blah", Age: 3},
			{Name: "Holliams", Age: 99},
			{Name: "Smart Guy", Age: 44},
		},
	})
}

func TestSelectWhereOrderBy(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 10,
		SelectableCols:  []string{"age", "name"},
		Where:           map[string]string{"iq": "> ?"},
		OrderableCols:   []string{"iq"},
	}
	q := Query{
		Page:      1,
		WhereArgs: map[string]interface{}{"iq": 80},
		OrderBy:   []string{"iq asc"},
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{Name: "Blah", Age: 3},
			{Name: "Meh", Age: 77},
			{Name: "Test Dude", Age: 7},
		},
	})
}

func TestSearchAndWhere(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		DefaultPageSize: 10,
		Where:           map[string]string{"iq": "> ?", "name": "like ?", "age": "> 0"},
		OrderableCols:   []string{"iq"},
	}
	q := Query{
		Page:      1,
		WhereArgs: map[string]interface{}{"iq": 80},
		OrderBy:   []string{"iq desc"},
		Search:    "%h%",
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
			{ID: 5, Name: "Blah", Age: 3, IQ: 100},
		},
	})

}

func TestFilterFunc(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		FilterFunc: func(db *gorm.DB, query Query) *gorm.DB {
			return db.Where("name NOT LIKE ?", "%dude%")
		},
	}
	q := Query{
		Page: 1,
	}

	testPagination(t, db, c, q, [][]dbModel{
		{
			{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
			{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
			{ID: 4, Name: "Meh", Age: 77, IQ: 120},
			{ID: 5, Name: "Blah", Age: 3, IQ: 100},
			{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
			{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
		},
	})
}

func TestInvalidQueryPage(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{}
	q := Query{
		Page: 0,
	}
	var results []dbModel
	_, err := Do(db, c, q, &results)
	assert.Error(t, err)
}

func TestBadWhere(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{}
	q := Query{
		Page:      1,
		WhereArgs: map[string]interface{}{"age": 7},
	}
	var results []dbModel
	_, err := Do(db, c, q, &results)
	assert.Error(t, err)
}

func TestBadSelect(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		SelectableCols: []string{"id"},
	}
	q := Query{
		Page:   1,
		Select: []string{"age"},
	}
	var results []dbModel
	_, err := Do(db, c, q, &results)
	assert.Error(t, err)
}

func TestBadOrderBy(t *testing.T) {
	db, f := setup(t)
	defer f()

	c := Config{
		OrderableCols: []string{"id"},
	}
	q := Query{
		Page:    1,
		OrderBy: []string{"age"},
	}
	var results []dbModel
	_, err := Do(db, c, q, &results)
	assert.Error(t, err)
}

func TestPatchLikeQuery(t *testing.T) {
	c := Config{
		Where: map[string]string{"name": "like ?", "id": "= ?"},
	}
	q := Query{
		WhereArgs: map[string]interface{}{"name": "bob", "id": 38, "bogus": "blah"},
		Search:    "yodda",
	}
	PatchLikeQuery(&c, &q, true, false)
	assert.Equal(t, 3, len(q.WhereArgs))          // No field was added or removed.
	assert.Equal(t, "%bob", q.WhereArgs["name"])  // Name was patched.
	assert.Equal(t, "blah", q.WhereArgs["bogus"]) // Not patched (does not match).
	assert.Equal(t, 38, q.WhereArgs["id"])        // Not patched (not string).
	assert.Equal(t, "%yodda", q.Search)           // Search is always patched.

	q = Query{
		WhereArgs: map[string]interface{}{"name": "bob", "id": 38, "bogus": "blah"},
		Search:    "yodda",
	}
	PatchLikeQuery(&c, &q, false, true)
	assert.Equal(t, 3, len(q.WhereArgs))          // No field was added or removed.
	assert.Equal(t, "bob%", q.WhereArgs["name"])  // Name was patched.
	assert.Equal(t, "blah", q.WhereArgs["bogus"]) // Not patched (does not match).
	assert.Equal(t, 38, q.WhereArgs["id"])        // Not patched (not string).
	assert.Equal(t, "yodda%", q.Search)           // Search is always patched.

	// Calling it again does not add extra "%"s.
	PatchLikeQuery(&c, &q, true, true)
	assert.Equal(t, 3, len(q.WhereArgs))          // No field was added or removed.
	assert.Equal(t, "bob%", q.WhereArgs["name"])  // Name was patched.
	assert.Equal(t, "blah", q.WhereArgs["bogus"]) // Not patched (does not match).
	assert.Equal(t, 38, q.WhereArgs["id"])        // Not patched (not string).
	assert.Equal(t, "yodda%", q.Search)           // Search is always patched.
}

func testPagination(t *testing.T, db *gorm.DB, c Config, q Query, resultsPerPage [][]dbModel) {
	var results [][]dbModel

	for {
		var local []dbModel
		res, err := Do(db, c, q, &local)
		if err != nil {
			t.Fatal(err)
		}
		if res.Error != nil {
			t.Fatal(res.Error)
		}
		if res.RowsAffected == 0 {
			break
		}
		results = append(results, local)
		q.Page++
	}
	assert.Equal(t, resultsPerPage, results)
}

type dbModel struct {
	ID   int64
	Name string
	Age  int16
	IQ   int32
}

var testData = []dbModel{
	{ID: 1, Name: "Don Jr", Age: 46, IQ: 1},
	{ID: 2, Name: "Potranka", Age: 44, IQ: 80},
	{ID: 3, Name: "Test Dude", Age: 7, IQ: 200},
	{ID: 4, Name: "Meh", Age: 77, IQ: 120},
	{ID: 5, Name: "Blah", Age: 3, IQ: 100},
	{ID: 6, Name: "Holliams", Age: 99, IQ: 50},
	{ID: 7, Name: "Smart Guy", Age: 44, IQ: 30},
}

func createDB() (*gorm.DB, func()) {
	// Create temp storage for sqlite
	tmpfile, err := ioutil.TempFile("", "page_test")
	if err != nil {
		panic(err)
	}
	dbName := tmpfile.Name()
	if err = tmpfile.Close(); err != nil {
		panic(err)
	}

	// side effect: db is now in model.M
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		panic(err)
	}
	gdb, err := gorm.Open("sqlite3", db)
	if err != nil {
		panic(err)
	}
	return gdb, func() { os.Remove(dbName) }
}

func setup(t *testing.T) (*gorm.DB, func()) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	gdb, f := createDB()
	res := gdb.AutoMigrate(&dbModel{})
	if err := res.Error; err != nil {
		t.Fatal(err)
	}
	for i, d := range testData {
		res = gdb.Create(&d)
		if err := res.Error; err != nil {
			t.Fatalf("error creating record %d:%s", i, err)
		}
	}
	return gdb, f
}
