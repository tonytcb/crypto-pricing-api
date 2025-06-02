package domain

import "strings"

type Pair struct {
	From Currency
	To   Currency
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
