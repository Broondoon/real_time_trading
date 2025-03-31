package objects

import "github.com/google/uuid"

type Pair struct {
	ID1 string `json:"id1"`
	ID2 string `json:"id2"`
}

func (p *Pair) String() string {
	return p.ID1 + "," + p.ID2
}

type FloatPair struct {
	FloatVal float64    `json:"floatVal"`
	IDVal    *uuid.UUID `json:"idVal"`
}

type StringPair struct {
	StringVal string     `json:"stringVal"`
	IDVal     *uuid.UUID `json:"idVal"`
}
