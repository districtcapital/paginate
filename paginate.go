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
	// no restrictions. If SelectableCols contains the same column name as
	// FilterCols, FilterCols wins.
	SelectableCols []string

	// FilterCols is a list of columns to suppress from the final output. An empty
	// list means all columns will show in the output. If FilterCols contains the
	// same column name as SelectableCols, FilterCols wins.
	FilterCols []string

	// FilterFunc pre-configures the query in a way that expands or restricts
	// the query. It is applied *before* the final GORM query is built.
	FilterFunc func(db *gorm.DB, query Query)

	// PageSize is the number of items to return per page. If zero,
	// defaultPageSize will be used. The page size is futher constrained by
	// maxPageSize.
	PageSize int

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

	// Page is the page to return, assuming the configured page size.
	// Pages start at 1.
	Page int

	// OrderBy describes the columns to order by and optionally the mode ("ASC"
	// or "DESC"). If OrderBy is not whitelisted by Config.OrderableCols, an
	// error is returned.
	OrderBy []string
}

// Do performs the querying and pagination as described by Query, subject to
// the constraints of Config. It populates the results in 'results'.
func Do(c Config, q Query, results interface{}) error {
	return nil
}
