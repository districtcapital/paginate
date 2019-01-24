// Copyright District Capital Inc 2019
// All rights reserved.

package paginate

import (
	"fmt"
)

// Person is the database object stored in the Persons table.
type Person struct {
	ID   uint
	Name string
	Age  int16
}

func Example() {
	db, f := createDB()
	defer f()

	if res := db.AutoMigrate(&Person{}); res.Error != nil {
		panic(res.Error)
	}

	for i, p := range []Person{
		{1, "Bob Smith", 48},
		{2, "Joan Of Arc", 312},
		{3, "Morihei Ueshiba", 69},
		{4, "John Doe", 19},
		{5, "Silvio Santos", 99},
	} {
		res := db.Create(&p)
		if err := res.Error; err != nil {
			panic(fmt.Errorf("error creating record %d: %s", i, err))
		}
	}

	// Configure the query and page size.
	c := Config{
		DefaultPageSize: 3,
		Where:           map[string]string{"age": "> ?"},
		OrderableCols:   []string{"name"},
	}

	// The query with its parameters. Many/most of these may come from an
	// HTTP request or other user input.
	q := Query{
		Page:      1,
		WhereArgs: map[string]interface{}{"age": 21},
		OrderBy:   []string{"name asc"},
	}

	var results []Person

	// Get first page of results,
	res, err := Do(db, c, q, &results)
	if err != nil {
		panic(err)
	}
	if res.Error != nil {
		panic(res.Error)
	}

	// Here we could send 'results' over an HTTPS connection.
	fmt.Println(results)

	// User asked for the next page of results.
	q.Page = 2
	res, err = Do(db, c, q, &results)
	if err != nil {
		panic(err)
	}
	if res.Error != nil {
		panic(res.Error)
	}

	// Print (or send over the network) the last set of results.
	fmt.Println(results)
	// Output:
	// [{1 Bob Smith 48} {2 Joan Of Arc 312} {3 Morihei Ueshiba 69}]
	// [{5 Silvio Santos 99}]
}
