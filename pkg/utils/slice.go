package utils

func DeduplicateSlice[T comparable, Slice ~[]T](collection Slice) Slice {
	result := make(Slice, 0, len(collection))
	seen := make(map[T]bool, len(collection))

	for i := range collection {
		if seen[collection[i]] {
			continue
		}

		seen[collection[i]] = true
		result = append(result, collection[i])
	}

	return result
}
