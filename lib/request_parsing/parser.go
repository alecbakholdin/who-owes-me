package request_parsing

import (
	"cmp"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

type Parser struct {
	VenmoCol  int
	AmountCol int
	NoteCol   int
	Separator rune
	HasHeader bool
	MaxCol    int
}

func NewParser(venmoCol, amountCol, noteCol int, separator rune, hasHeader bool) *Parser {
	max := venmoCol
	if amountCol > max {
		max = amountCol
	}
	if noteCol > max {
		max = noteCol
	}

	return &Parser{
		VenmoCol:  venmoCol,
		AmountCol: amountCol,
		NoteCol:   noteCol,
		Separator: separator,
		MaxCol:    max,
	}
}

func (p *Parser) Parse(r io.Reader) ([]ParsedRequests, error) {
	requestMap := make(map[string]*ParsedRequests)
	reader := csv.NewReader(r)
	reader.Comma = p.Separator
	skip := p.HasHeader
	for {
		line, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		} else if len(line) == 0 {
			fmt.Println("breaking")
			break
		} else if skip {
			skip = false
			fmt.Println("skipping")
			continue
		} else if len(line) <= p.MaxCol {
			return nil, fmt.Errorf("expected at least %d cols but found %d for line %v", p.MaxCol+1, len(line), line)
		}

		venmo := line[p.VenmoCol]
		amt, err := p.parseAmount(line[p.AmountCol])
		if err != nil {
			return nil, err
		}
		note := line[p.NoteCol]

		key := fmt.Sprintf("%s_%.02f", note, amt)
		pr, ok := requestMap[key]
		if !ok {
			pr = &ParsedRequests{Amount: amt, Note: note}
			requestMap[key] = pr
		}
		pr.AddVenmo(venmo)
	}

	arr := []ParsedRequests{}
	for _, r := range requestMap {
		arr = append(arr, *r)
	}
	sort.Slice(arr, func(i, j int) bool {
		return cmp.Or(
			strings.Compare(arr[i].Note, arr[j].Note) < 0,
			arr[i].Amount < arr[i].Amount,
		)
	})
	return arr, nil
}

// parses one of the following:
// 1
// 1.23
// $1.23
// $ 1.23
// -1.23
// $-1.23
// $ -1.23
// $ (1.23)
// $(1.23)
func (p *Parser) parseAmount(str string) (float64, error) {
	str = strings.TrimSpace(strings.Replace(str, "$", "", 1))
	leftP := strings.IndexRune(str, '(')
	rightP := strings.IndexRune(str, ')')
	var negate float64 = 1
	if leftP >= 0 && rightP > leftP {
		str = strings.TrimSpace(str[leftP+1 : rightP])
		negate = -1
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return f, err
	}
	return negate * f, err
}

