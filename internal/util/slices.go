package util

func ConcatSlices[T any](slices ...[]T) []T {
	totalLength := 0
	for _, slice := range slices {
		totalLength += len(slice)
	}

	var i int

	result := make([]T, totalLength)
	for _, slice := range slices {
		i += copy(result[i:], slice)
	}

	return result

}
