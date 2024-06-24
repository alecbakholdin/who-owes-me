package request_parsing

type ParsedRequests struct {
	venmoBatches [][]string
	Amount       float64
	Note         string
}

// number of venmos in this request
func (p *ParsedRequests) Size() int {
	size := 0
	for _, b := range p.venmoBatches {
		size += len(b)
	}
	return size
}

func (p *ParsedRequests) AddVenmo(venmo string) {
	if p.venmoBatches == nil {
		p.venmoBatches = [][]string{{venmo}}
		return
	}
	for i, batch := range p.venmoBatches {
		if len(batch) < 10 {
			p.venmoBatches[i] = append(batch, venmo)
			return
		}
	}
	p.venmoBatches = append(p.venmoBatches, []string{venmo})
}
