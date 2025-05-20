package uuid

import "github.com/gofrs/uuid/v5"

func NewV7() uuid.UUID {
	var (
		uuidV7 uuid.UUID
		err    error
	)

	uuidV7, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil
	}

	return uuidV7
}
