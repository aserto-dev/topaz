package lox

// ContainsPtr similar to lo.Contains for slices of pointers.
func ContainsPtr[T comparable](collection []*T, element *T) bool {
	for _, item := range collection {
		if *item == *element {
			return true
		}
	}

	return false
}

// DifferencePtr similar to lo.Difference but for slices of pointers.
func DifferencePtr[T comparable](list1, list2 []*T) ([]*T, []*T) {
	left := []*T{}
	right := []*T{}

	seenLeft := map[T]struct{}{}
	seenRight := map[T]struct{}{}

	for _, elem := range list1 {
		seenLeft[*elem] = struct{}{}
	}

	for _, elem := range list2 {
		seenRight[*elem] = struct{}{}
	}

	for _, elem := range list1 {
		if _, ok := seenRight[*elem]; !ok {
			left = append(left, elem)
		}
	}

	for _, elem := range list2 {
		if _, ok := seenLeft[*elem]; !ok {
			right = append(right, elem)
		}
	}

	return left, right
}
