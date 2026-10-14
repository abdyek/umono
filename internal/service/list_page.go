package service

// listPageSize is the number of items an admin sidebar shows at a time.
const listPageSize = 50

// pageOf takes the items read from offset with limit listPageSize+1 and
// returns at most listPageSize of them with the offset of the next page, or 0
// when there is no next page.
func pageOf[T any](items []T, offset int) ([]T, int) {
	if len(items) > listPageSize {
		return items[:listPageSize], offset + listPageSize
	}
	return items, 0
}
