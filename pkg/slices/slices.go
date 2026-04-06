package slices

import "fmt"

func Map[T any, U any](s []T, f func(T) (U, error)) ([]U, error) {
	var err error
	result := make([]U, len(s))

	for i, v := range s {
		result[i], err = f(v)
		if err != nil {
			return nil, fmt.Errorf("slice map: %w", err)
		}
	}

	return result, nil
}
