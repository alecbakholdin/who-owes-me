package request_parsing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsedRequests(t *testing.T) {
	t.Run("adding one venmo", func(t *testing.T) {
		pr := ParsedRequests{}
		pr.AddVenmo("venmo1")

		assert.Equal(t, 1, pr.Size())
		assert.Equal(t, "venmo1", pr.VenmoBatches[0][0])
	})

	t.Run("adding more than one batch of venmos", func(t *testing.T) {
		pr := ParsedRequests{}
		for i := range 30 {
			pr.AddVenmo(fmt.Sprintf("venmo%d", i))
		}

		assert.Equal(t, 30, pr.Size())
		for i := range 3 {
			for j := range 10 {
				assert.Equal(t, fmt.Sprintf("venmo%d", i*10+j), pr.VenmoBatches[i][j])
			}
		}
	})
}
