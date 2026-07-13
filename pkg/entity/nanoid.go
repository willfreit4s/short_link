// Package entity defines the ID type and its validation logic.
package entity

import (
	"errors"
	"regexp"

	"github.com/jaevor/go-nanoid"
)

const IDLength = 12
const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-"

var generateNanoID func() string
var idRegex = regexp.MustCompile(`^[0-9A-Za-z_-]{12}$`)

var ErrInvalidID = errors.New("identifier with an invalid format or size")

type NanoID string

func init() {
	var err error
	generateNanoID, err = nanoid.CustomASCII(Alphabet, IDLength)
	if err != nil {
		panic("failure to initialize NanoID generator: " + err.Error())
	}
}

func (id NanoID) String() string {
	return string(id)
}

func NewNanoID() NanoID {
	return NanoID(generateNanoID())
}

func ParseID(raw string) (NanoID, error) {
	if len(raw) != IDLength {
		return "", ErrInvalidID
	}

	if !idRegex.MatchString(raw) {
		return "", ErrInvalidID
	}

	return NanoID(raw), nil
}
