package engine

// AddQuery добавляет фразу для поиска в движок.
// Проверяет фразу на уникальность в рамках движка.
func (i *impl) AddQuery(query string) {
	var found bool
	for _, savedQuery := range i.queries {
		if query == savedQuery {
			found = true
			break
		}
	}

	if !found {
		i.queries = append(i.queries, query)
	}
}
