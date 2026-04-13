package pqutil

import (
	"errors"

	"github.com/lib/pq"
)

func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
