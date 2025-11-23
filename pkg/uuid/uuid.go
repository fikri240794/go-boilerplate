package uuid

import "github.com/gofrs/uuid/v5"

type uuidGenerator func() (uuid.UUID, error)

var defaultUUIDGenerator uuidGenerator = uuid.NewV7

func NewV7() uuid.UUID {
	return newV7WithGenerator(defaultUUIDGenerator)
}

func newV7WithGenerator(generator uuidGenerator) uuid.UUID {
	var (
		uuidV7 uuid.UUID
		err    error
	)

	uuidV7, err = generator()
	if err != nil {
		return uuid.Nil
	}

	return uuidV7
}
