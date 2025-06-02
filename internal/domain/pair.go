package domain

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
