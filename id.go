package zinc

import (
	"encoding/hex"
	"errors"
	"math/rand"
)

var (
	ErrInvalidUid = errors.New("uid must be a 40 char hex string")
)

// Uid is a 160bit number that identifies objects in a cluster
type Uid [20]byte

func (id Uid) String() string {
	return hex.EncodeToString(id[:])
}

// RandomDevUid reads random bytes generates uid with the systems random device
func RandomDevUid() {}

// ParseUid converts a uid string to `Uid` type
func ParseUid(data string) (Uid, error) {
	if len(data) != 40 {
		return Uid{}, ErrInvalidUid
	}

	duid, err := hex.DecodeString(data)
	if err != nil {
		return Uid{}, err
	}

	var uid Uid
	if n := copy(uid[:], duid); n != 20 {
		return Uid{}, ErrInvalidUid
	}

	return uid, nil
}

// RandomUid generates a uid with random numbers
func RandomUid() Uid {
	var uid Uid
	for i := 0; i < 20; i++ {
		uid[i] = byte(rand.Intn(256))
	}
	return uid
}
