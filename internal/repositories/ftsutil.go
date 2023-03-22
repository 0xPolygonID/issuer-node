package repositories

import "strings"

// fullTextSearchQuery accepts a query with a list of words and returns a tsquery that includes words that
// begin or contains that words. operator is used to pass an operator between words.
// https://www.postgresql.org/docs/current/datatype-textsearch.html#DATATYPE-TSQUERY
func fullTextSearchQuery(query string, operator string) string {
	words := strings.Split(strings.ReplaceAll(query, ",", " "), " ")
	terms := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		terms = append(terms, "("+word+":* | "+word+")")
	}
	return strings.Join(terms, operator)
}

// getDIDFromQuery searches for words that begin with "did:" and returns the first occurrence. Empty string otherwise
func getDIDFromQuery(query string) string {
	words := strings.Split(strings.ReplaceAll(query, ",", " "), " ")
	for _, word := range words {
		if strings.HasPrefix(word, "did:") {
			return word
		}
	}
	return ""
}
