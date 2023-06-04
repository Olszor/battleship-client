package app

func Filter[T any](data []T, f func(T) bool) []T {
	r := make([]T, 0, len(data))

	for _, element := range data {
		if f(element) {
			r = append(r, element)
		}
	}

	return r
}

func Map[T, U any](data []T, f func(T) U) []U {
	r := make([]U, 0, len(data))

	for _, element := range data {
		r = append(r, f(element))
	}

	return r
}
