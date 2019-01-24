# paginate: Simple paging for GORM

## Install

`go get -u github.com/districtcapital/paginate`

## Usage

```Go
// Config configures the base query -- what can be done. In this example the
// calling user can filter by age greater than a parameter and id equal to
// some other parameter and order by column iq.
c := Config{
  DefaultPageSize: 10,
  Where:           map[string]string{"age": "> ?", "id": "= ?"},
  OrderableCols:   []string{"iq"},
  SelectableCols:  []string{"age", "iq", "name"},
}
// Query is the user input, possibly coming from an HTTP request. Not all
// configured parameters above need to be present. In fact, the only required
// parameter is Page.
q := Query{
  Page:      1,
  WhereArgs: map[string]interface{}{"age":40, "id": 1234},
  OrderBy:   []string{"iq"},
}

// db, err := gorm.Open(...)

var results []interface{}

// Execute the query c as bound by parameters q on the gorm db and output results.
res, err := Do(db, c, q, &results)
```

## Contributions

Pull requests are accepted so long they add a feature that is generic enough to benefit many as opposed to being something only one person/company will deem relevant (and yes, that's a subjective call).

Pull requests _must_ contain good, solid tests or they will be rejected.

## License

This library is licensed under the MIT License. See file [LICENSE] for details.

No part of this license grants anyone the right to use the name of District Capital Inc in any form, neither as endorsement or to promote any product or service.
