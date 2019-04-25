// Copyright District Capital Inc 2019
// All rights reserved.

package paginate

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
)

func build(db *gorm.DB, c *Config, q *Query) (*gorm.DB, error) {
	s, err := selectCols(c, q)
	if err != nil {
		return nil, err
	}
	if s != "" {
		db = db.Select(s)
	}
	w, wa, err := where(c, q)
	if err != nil {
		return nil, err
	}
	if w != "" {
		db = db.Where(w, wa...)
	}
	o, err := orderBy(c, q)
	if err != nil {
		return nil, err
	}
	if o != "" {
		db = db.Order(o)
	}
	if q.Page <= 0 {
		return nil, fmt.Errorf("invalid page: %d", q.Page)
	}
	pageSize := pageSize(c, q)
	offset := uint64(pageSize) * uint64(q.Page-1)
	if c.FilterFunc != nil {
		db = c.FilterFunc(db, *q)
	}
	return db.Offset(offset).Limit(pageSize), nil
}

func pageSize(c *Config, q *Query) uint16 {
	if c.DefaultPageSize == 0 {
		c.DefaultPageSize = defaultPageSize
	}
	if c.MaxPageSize == 0 {
		c.MaxPageSize = maxPageSize
	}
	pageSize := q.PageSize
	if pageSize == 0 {
		return c.DefaultPageSize
	}
	if pageSize > c.MaxPageSize {
		return c.MaxPageSize
	}
	return pageSize
}

// orderBy builds the ORDER BY clause.
func orderBy(c *Config, q *Query) (string, error) {
	var buf bytes.Buffer

Outer:
	for _, o := range q.OrderBy {
		oo := strings.ToLower(strings.TrimSpace(o))
		ob := strings.Split(oo, " ")
		if len(ob[0]) == 0 {
			// We got an empty order by. Nothing to do.
			continue
		}
		if len(ob) > 2 {
			return "", fmt.Errorf("invalid order_by clause %q", o)
		}
		if len(ob) == 2 {
			if ob[1] != "asc" && ob[1] != "desc" {
				return "", fmt.Errorf("invalid sort direction in order_by clause %q", o)
			}
		}
		for _, oc := range c.OrderableCols {
			if strings.EqualFold(ob[0], oc) {
				pad(&buf, ", ")
				buf.WriteString(oo)
				continue Outer
			}
		}
		return "", fmt.Errorf("query cannot order by field %q", o)
	}
	return buf.String(), nil
}

// selectCols builds the SELECT clause.
func selectCols(c *Config, q *Query) (string, error) {
	var buf bytes.Buffer

	// No mention of a selectable column means all columns are allowed.
	if len(c.SelectableCols) == 0 {
		// If the query did not specify, select everything.
		if len(q.Select) == 0 {
			return "*", nil
		}
	}

Outer:
	for _, ss := range q.Select {
		s := strings.ToLower(strings.TrimSpace(ss))
		if len(s) == 0 {
			// We got an empty select. Nothing to do.
			continue
		}
		// If we don't restrict any columns, whatever comes can be added.
		if len(c.SelectableCols) == 0 {
			pad(&buf, ", ")
			buf.WriteString(s)
		} else {
			for _, sc := range c.SelectableCols {
				if strings.EqualFold(s, sc) {
					pad(&buf, ", ")
					buf.WriteString(s)
					continue Outer
				}
			}
			return "", fmt.Errorf("query cannot select column %q", s)
		}
	}
	// If we did not select anything, then we select *everything* that *can* be
	// selected (but for efficiency not "*").
	if buf.Len() == 0 {
		return strings.ToLower(strings.Join(c.SelectableCols, ", ")), nil
	}
	return buf.String(), nil
}

// where builds the WHERE clause.
func where(c *Config, q *Query) (string, []interface{}, error) {
	var args []interface{}

	// Are we disallowing Search, but Search is requested?
	if c.DisallowSearchTerm && q.Search != "" {
		return "", nil, fmt.Errorf("search term is disallowed by config")
	}

	// Maps are unsorted so we sort the keys to ensure testable results.
	keys := make([]string, 0, len(q.WhereArgs))
	valuesWithNewKeys := make(map[string]interface{})
	for k, v := range q.WhereArgs {
		kk := strings.ToLower(strings.TrimSpace(k))
		keys = append(keys, kk)
		valuesWithNewKeys[kk] = v
	}
	sort.Strings(keys)

	var buf bytes.Buffer

	// We reject WhereArg keys that are not in Where keys.
	for _, k := range keys {
		if _, found := c.Where[k]; !found {
			return "", nil, fmt.Errorf("where argument %q not allowed", k)
		}
		pad(&buf, " AND ")
		buf.WriteString(k)
		buf.WriteString(" ")
		buf.WriteString(c.Where[k])
		args = append(args, valuesWithNewKeys[k])
	}

	// If there is no search term, we're done.
	if q.Search == "" {
		return buf.String(), args, nil
	}

	// When Search is on, we apply the Search to all LIKE queries.
	keys = likeClauses(c)

	var orBuf bytes.Buffer
	for _, k := range keys {
		pad(&orBuf, " OR ")
		orBuf.WriteString(k)
		orBuf.WriteString(" ")
		orBuf.WriteString(c.Where[k])
		args = append(args, q.Search)
	}

	and := buf.Len() > 0
	or := orBuf.Len() > 0
	if and && or {
		buf.WriteString(" AND (")
		buf.ReadFrom(&orBuf)
		buf.WriteString(")")
	} else if or {
		buf.ReadFrom(&orBuf)
	}
	return buf.String(), args, nil
}

// likeClauses returns the sorted keys of all Where clauses that have a "LIKE"
// or "like" in them.
func likeClauses(c *Config) []string {
	var keys []string
	for k, v := range c.Where {
		if strings.Contains(v, "like") || strings.Contains(v, "LIKE") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// pad adds s to buf if the buffer is not empty.
func pad(buf *bytes.Buffer, s string) {
	if buf.Len() > 0 {
		buf.WriteString(s)
	}
}
