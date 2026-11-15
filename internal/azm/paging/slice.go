package paging

import (
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type (
	KeyComparer[T any] func([]string, T) bool
	KeyMapper[T any]   func(T) []string
)

type Result[T any] struct {
	Items     []T
	NextToken string
}

func PaginateSlice[T any](
	s []T,
	size int32,
	token string,
	keyCount int,
	cmp KeyComparer[T],
	mapper KeyMapper[T],
) (*Result[T], error) {
	result := &Result[T]{}

	start := 0

	if token != "" {
		cursor, err := DecodeCursor(token)
		if err != nil {
			return result, err
		}

		if len(cursor.Keys) != keyCount {
			return result, derr.ErrInvalidCursor
		}

		_, start, _ = lo.FindIndexOf(s, func(o T) bool {
			return cmp(cursor.Keys, o)
		})

		if start == -1 {
			return result, derr.ErrInvalidCursor
		}
	}

	pageSize := lo.Min([]int{int(size), len(s) - start})
	end := start + pageSize

	var next *string

	if end < len(s) {
		cursor := &Cursor{Keys: mapper(s[end])}

		n, err := cursor.Encode()
		if err != nil {
			return result, errors.Wrap(err, "failed to encode cursor")
		}

		next = &n
	}

	result.Items = s[start:end]
	result.NextToken = lo.FromPtr(next)

	return result, nil
}
