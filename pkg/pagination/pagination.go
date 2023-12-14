package pagination

const defaultMaxResults = 50

// Filter is a struct that contains the pagination filter
type Filter struct {
	MaxResults uint
	Page       *uint
}

// NewFilter creates a new filter
func NewFilter(maxResults *uint, page *uint) *Filter {
	var maxR uint = defaultMaxResults
	if maxResults != nil {
		maxR = *maxResults
	}

	return &Filter{
		MaxResults: maxR,
		Page:       page,
	}
}

// GetLimit returns the limit for the query
func (f *Filter) GetLimit() uint {
	return f.MaxResults
}

// GetOffset returns the offset for the query
func (f *Filter) GetOffset() uint {
	return (*f.Page - 1) * f.MaxResults
}
