package utils

func DeduplicateSlice[T comparable, Slice ~[]T](collection Slice) Slice {
	result := make(Slice, 0, len(collection))
	seen := make(map[T]struct{}, len(collection))

	for i := range collection {
		if _, ok := seen[collection[i]]; ok {
			continue
		}

		seen[collection[i]] = struct{}{}
		result = append(result, collection[i])
	}

	return result
}
