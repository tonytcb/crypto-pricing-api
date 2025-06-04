package domain

import (
	"github.com/pkg/errors"
	"strings"
)

type Pair struct {
	From Currency
	To   Currency
}

func NewPairFromString(v string) (Pair, error) {
	const wantLength = 6

	if len(v) != wantLength {
		return Pair{}, errors.New("invalid pair length, expected 6 characters")
	}

	return NewPair(
		Currency(v[:3]),
		Currency(v[3:]),
	), nil
}

func NewPair(from, to Currency) Pair {
	return Pair{
		From: from,
		To:   to,
	}
}

func (p Pair) String() string {
	return strings.ToUpper(string(p.From + p.To))
}
