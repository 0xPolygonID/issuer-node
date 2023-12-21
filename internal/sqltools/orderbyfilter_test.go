package sqltools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderByFilters(t *testing.T) {
	for _, ts := range []struct {
		name     string
		filters  OrderByFilters
		expected string
	}{
		{
			name:     "empty",
			filters:  OrderByFilters{},
			expected: "",
		},
		{
			name: "one filter",
			filters: OrderByFilters{
				OrderByFilter{
					Field: "field1",
					Desc:  false,
				},
			},
			expected: "field1 ASC",
		},
		{
			name: "one filter desc",
			filters: OrderByFilters{
				OrderByFilter{
					Field: "field1",
					Desc:  true,
				},
			},
			expected: "field1 DESC",
		},
		{
			name: "two filters",
			filters: OrderByFilters{
				OrderByFilter{
					Field: "field1",
					Desc:  false,
				},
				OrderByFilter{
					Field: "field2",
					Desc:  true,
				},
			},
			expected: "field1 ASC, field2 DESC",
		},
		{
			name: "three filters",
			filters: OrderByFilters{
				OrderByFilter{
					Field: "field1",
					Desc:  false,
				},
				OrderByFilter{
					Field: "field2",
					Desc:  true,
				},
				OrderByFilter{
					Field: "field3",
					Desc:  false,
				},
			},
			expected: "field1 ASC, field2 DESC, field3 ASC",
		},
	} {
		t.Run(ts.name, func(t *testing.T) {
			filters := OrderByFilters{}
			for _, f := range ts.filters {
				err := filters.Add(f.Field, f.Desc)
				assert.NoError(t, err)
			}
			assert.Equal(t, ts.expected, filters.String())
		})
	}

	t.Run("duplicated filters", func(t *testing.T) {
		filters := OrderByFilters{}
		err := filters.Add("field1", false)
		assert.NoError(t, err)
		err = filters.Add("field1", false)
		assert.Error(t, err)
	})
}
