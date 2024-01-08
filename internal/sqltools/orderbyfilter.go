package sqltools

import (
	"errors"
	"strings"
)

// SQLFieldName is an alias for string and is used to define order by filter constants
type SQLFieldName string

// OrderByFilter represents a filter over a field with an specific order (ASC(false) or DESC (true))
type OrderByFilter struct {
	Field     SQLFieldName
	Desc      bool
	NullsLast bool
}

// OrderByFilters is a collection of OrderByFilter with some handy methods to add order filters
// and generate an SQL LIKE ORDER BY clause
type OrderByFilters []OrderByFilter

// Add adds a new OrderByFilter to the collection. If the field already exists, it returns an error
func (s *OrderByFilters) Add(f SQLFieldName, desc bool) error {
	return s.add(f, desc, false)
}

// AddWithNullsLast adds a new OrderByFilter to the collection with the NULLS LAST modifier. If the field already exists, it returns an error
func (s *OrderByFilters) AddWithNullsLast(f SQLFieldName, desc bool) error {
	return s.add(f, desc, true)
}

// String returns an SQL LIKE ORDER BY clause
func (s *OrderByFilters) String() string {
	var sortFields []string
	for _, sortBy := range *s {
		s := string(sortBy.Field)
		if sortBy.Desc {
			s += " DESC"
		} else {
			s += " ASC"
		}
		if sortBy.NullsLast {
			s += " NULLS LAST"
		}
		sortFields = append(sortFields, s)
	}
	return strings.Join(sortFields, ", ")
}

func (s *OrderByFilters) add(f SQLFieldName, desc bool, withNullsLast bool) error {
	for _, v := range *s {
		if v.Field == f {
			return errors.New("sql sort filter field already exists")
		}
	}
	*s = append(*s, OrderByFilter{Field: f, Desc: desc, NullsLast: withNullsLast})
	return nil
}
