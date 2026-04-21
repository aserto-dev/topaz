package validator

import (
	"errors"

	dsc "github.com/aserto-dev/topaz/api/directory/v4"
)

var ErrPaginationSize = errors.New("pagination size must be >= 1 and <= 100")

func PaginationRequest(msg *dsc.PaginationRequest) error {
	if msg == nil {
		return nil
	}

	if msg.GetSize() < 1 || msg.GetSize() > 100 {
		return ErrPaginationSize
	}

	return nil
}
