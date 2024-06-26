package request_parsing

type ParsedRequests struct {
	VenmoBatches [][]string
	Amount       float64
	Note         string
}

// number of venmos in this request
func (p *ParsedRequests) Size() int {
	size := 0
	for _, b := range p.VenmoBatches {
		size += len(b)
	}
	return size
}

func (p *ParsedRequests) AddVenmo(venmo string) {
	if p.VenmoBatches == nil {
		p.VenmoBatches = [][]string{{venmo}}
		return
	}
	for i, batch := range p.VenmoBatches {
		if len(batch) < 10 {
			p.VenmoBatches[i] = append(batch, venmo)
			return
		}
	}
	p.VenmoBatches = append(p.VenmoBatches, []string{venmo})
}
