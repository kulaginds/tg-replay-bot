package engine

import (
	"strings"
)

// HasQueries ищет фразы поиска в тексте.
func (i *impl) HasQueries(text string) bool {
	var hasQueries bool

	for _, query := range i.queries {
		if strings.Contains(text, query) {
			hasQueries = true
			break
		}
	}

	return hasQueries
}
