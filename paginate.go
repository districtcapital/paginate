// Copyright District Capital Inc 2019
// All rights reserved.

// Package paginate performs search, filtering and pagination for GORM
package paginate

import (
	"github.com/jinzhu/gorm"
)

// Config configures a search and pagination request.
type Config struct {
	// SelectableCols restricts which columns may be selected. An empty list means
	// no restrictions.
	SelectableCols []string

	// FilterFunc pre-configures the query in a way that expands or restricts
	// the query. It is applied *before* the final GORM query is built.
	FilterFunc func(db *gorm.DB, query Query) *gorm.DB

	// MaxPageSize is the maximum number of elements a query can request in one
	// page. If MaxPageSize is not set, it defaults to maxPageSize.
	MaxPageSize uint16

	// DefaultPageSize is the default page size to use if the user does not
	// request another page size. If DefaultPageSize is not set, a reasonable
	// default (defaultPageSize) is used.
	DefaultPageSize uint16

	// OrderableCols is a list of all columns that can be ordered by.
	OrderableCols []string

	// Where describes the possible where clauses that are optionally
	// matched against WhereArgs in the Query.
	// E.g. {"id": "> ?", "doc_age": "< ?"} would match with WhereArgs
	// {"id": 32, "doc_age": 128} but not with {"user_id": 1, "age": 7}
	Where map[string]string
}

// Query declares a query instance, used for querying a model subject to the
// constraints of the Config.
type Query struct {
	// Select represents columns being selected. An empty list means "*". If a
	// column name is not whitelisted by SelectableCols, an error is returned.
	Select []string

	// WhereArgs describes the arguments used in the where clause. The keys are
	// matched against Config.Where as chained AND clauses. For example, if Where
	// is {"name": "LIKE %?%", "iq": "< ?"} and WhereArgs is
	// {"name": "Trump", "iq": 100} the final where clause would be
	// WHERE name like %Trump% AND iq < 100
	WhereArgs map[string]interface{}

	// PageSize is the number of items to return per page. If zero,
	// the Config's DefaultPageSize will be used. The page size is futher
	// constrained by config.MaxPageSize.
	PageSize uint16

	// Page is the page to return, assuming the configured page size.
	// Pages start at 1.
	Page uint32

	// OrderBy describes the columns to order by and optionally the mode ("ASC"
	// or "DESC"). If OrderBy is not whitelisted by Config.OrderableCols, an
	// error is returned.
	OrderBy []string
}

const (
	defaultPageSize = 25
	maxPageSize     = 1000
)

// Do performs the querying and pagination as described by Query, subject to
// the constraints of Config. It populates the results in 'results'.
// An error-less return does not mean the query succeeded, it only means the
// query builder succeeded -- one must also check the Error field in gorm.DB.
func Do(db *gorm.DB, c Config, q Query, results interface{}) (*gorm.DB, error) {
	var err error
	db, err = build(db, &c, &q)
	if err != nil {
		return nil, err
	}
	return db.Find(results), nil
}
