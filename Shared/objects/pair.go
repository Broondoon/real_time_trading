package objects

type Pair struct {
	ID1 string `json:"id1"`
	ID2 string `json:"id2"`
}

func (p *Pair) String() string {
	return p.ID1 + "," + p.ID2
}
