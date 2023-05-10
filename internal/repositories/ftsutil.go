package repositories

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	onlyLettersAndNumbers = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	didCharacters         = regexp.MustCompile(`[^a-zA-Z0-9:]+`)
)

// fullTextSearchQuery accepts a query with a list of words and returns a tsquery that includes words that
// begin or contains that words. operator is used to pass an operator between words.
// https://www.postgresql.org/docs/current/datatype-textsearch.html#DATATYPE-TSQUERY
func fullTextSearchQuery(query string, operator string) string {
	query = onlyLettersAndNumbers.ReplaceAllString(query, " ")
	words := strings.Split(query, " ")
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

func tokenizeQuery(query string) []string {
	words := strings.Split(strings.ReplaceAll(query, ",", " "), " ")
	terms := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" && !inArray(word, terms) {
			terms = append(terms, word)
		}
	}
	return terms
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

// buildPartialQueryDidLikes accepts a list of words and returns a SQL LIKE sentence to match this words against a field.
// Example:
// field:= dbfield; words := []string{"word1", "word2"}; operator := "OR"
//
// returns "dbfield ILIKE '%word1%' OR did ILIKE '%word2%'"
func buildPartialQueryDidLikes(field string, words []string, cond string) string {
	conditions := make([]string, 0, len(words))
	for _, word := range words {
		if word != "" {
			conditions = append(conditions, fmt.Sprintf("%s ILIKE '%%%s%%'", field, escapeDID(word)))
		}
	}
	return strings.Join(conditions, " "+cond+" ")
}

func buildPartialQueryLikes(field string, cond string, first int, n int) string {
	conditions := make([]string, 0, n)
	current := first
	for i := 0; i < n; i++ {
		conditions = append(conditions, fmt.Sprintf("%s ILIKE '%%' || $%d || '%%'", field, current))
		current++
	}
	return strings.Join(conditions, " "+cond+" ")
}

func escapeDID(s string) string {
	return didCharacters.ReplaceAllString(s, "")
}

func inArray(needle string, haystack []string) bool {
	for _, word := range haystack {
		if needle == word {
			return true
		}
	}
	return false
}
