package pagination

const defaultMaxResults = 1000

// Filter is a struct that contains the pagination filter
type Filter struct {
	MaxResults uint
	Page       *uint
}

// NewFilter creates a new filter
func NewFilter(maxResults *uint, page *uint) *Filter {
	f := &Filter{
		MaxResults: defaultMaxResults,
		Page:       page,
	}
	if maxResults != nil {
		f.MaxResults = *maxResults
	}
	return f
}

// GetLimit returns the limit for the query
func (f *Filter) GetLimit() uint {
	if f.MaxResults == 0 {
		return defaultMaxResults
	}
	return f.MaxResults
}

// GetOffset returns the offset for the query
func (f *Filter) GetOffset() uint {
	if f.Page == nil {
		return 0
	}
	return (*f.Page - 1) * f.MaxResults
}
