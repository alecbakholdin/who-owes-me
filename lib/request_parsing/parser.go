package request_parsing

import "io"

type Parser struct {
	VenmoCol  int
	AmountCol int
	NoteCol   int
	Separator string
	HasHeader bool
}

func NewParser(venmoCol, amountCol, noteCol int, separator string, hasHeader bool) *Parser {
	return &Parser{
		VenmoCol:  venmoCol,
		AmountCol: amountCol,
		NoteCol:   noteCol,
		Separator: separator,
	}
}

func (p *Parser) Parse(r io.Reader) {
	
}
