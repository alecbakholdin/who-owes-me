package request_parsing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	t.Run("parser exits normally with empty line", func(t *testing.T) {
		r := strings.NewReader("\n\n")
		parser := NewParser(0, 1, 2, '\t', true)
		arr, err := parser.Parse(r)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(arr))
	})
	t.Run("parser exits normally with empty string", func(t *testing.T) {
		r := strings.NewReader("")
		parser := NewParser(0, 1, 2, '\t', true)
		arr, err := parser.Parse(r)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(arr))
	})
	t.Run("single row", func(t *testing.T) {
		t.Run("basic amount and quotes", func(t *testing.T) {
			r := strings.NewReader("venmo\t1.23\t\"note\"")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 1.23, Note: "note"}, arr[0])
		})
		t.Run("negative amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t-1.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: -1.23, Note: "note"}, arr[0])
		})
		t.Run("dollar amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t$ 1.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 1.23, Note: "note"}, arr[0])
		})
		t.Run("negative dollar amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t$ -1.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: -1.23, Note: "note"}, arr[0])
		})
		t.Run("parenthetic negative dollar amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t$(1.23)\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: -1.23, Note: "note"}, arr[0])
		})
		t.Run("no leading number parenthetic amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t$(.23)\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: -0.23, Note: "note"}, arr[0])
		})
		t.Run("no leading number positive amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t$.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 0.23, Note: "note"}, arr[0])
		})
		t.Run("no leading number non-dollar positive amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 0.23, Note: "note"}, arr[0])
		})
		t.Run("integer amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t1\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 1, Note: "note"}, arr[0])
		})
	})
	t.Run("parser fails to parse faulty amounts", func(t *testing.T) {
		t.Run("empty float", func(t *testing.T) {
			r := strings.NewReader("venmo\t\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			_, err := parser.Parse(r)
			assert.NotNil(t, err)
		})
		t.Run("invalid float", func(t *testing.T) {
			r := strings.NewReader("venmo\t1.2.3\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			_, err := parser.Parse(r)
			assert.NotNil(t, err)
		})
		t.Run("unclosed parentheses financial", func(t *testing.T) {
			r := strings.NewReader("venmo\t$(1.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			_, err := parser.Parse(r)
			assert.NotNil(t, err)
		})
		t.Run("unclosed parentheses financial", func(t *testing.T) {
			r := strings.NewReader("venmo\t$1.23)\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			_, err := parser.Parse(r)
			assert.NotNil(t, err)
		})
		t.Run("spaces in float", func(t *testing.T) {
			r := strings.NewReader("venmo\t1 . 2\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			_, err := parser.Parse(r)
			assert.NotNil(t, err)
		})
		t.Run("double dollar sign", func(t *testing.T) {
			r := strings.NewReader("venmo\t$$1.23\tnote")
			parser := NewParser(0, 1, 2, '\t', false)
			_, err := parser.Parse(r)
			assert.NotNil(t, err)
		})
	})
	t.Run("multiple rows", func(t *testing.T) {
		t.Run("same note and amount", func(t *testing.T) {
			r := strings.NewReader("venmo\t1.23\tnote\nvenmo-2\t1.23\tnote\n")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo", "venmo-2"}}, Amount: 1.23, Note: "note"}, arr[0])
		})
		t.Run("different note", func(t *testing.T) {
			r := strings.NewReader("venmo\t1.23\tnote\nvenmo-2\t1.23\tnote-diff\n")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 1.23, Note: "note"}, arr[0])
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo-2"}}, Amount: 1.23, Note: "note-diff"}, arr[1])
		})
		t.Run("different amt", func(t *testing.T) {
			r := strings.NewReader("venmo\t1.23\tnote\nvenmo-2\t1.24\tnote\n")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 1.23, Note: "note"}, arr[0])
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo-2"}}, Amount: 1.24, Note: "note"}, arr[1])
		})
		t.Run("different both", func(t *testing.T) {
			r := strings.NewReader("venmo\t1.23\tnote\nvenmo-2\t1.24\tnote-diff\n")
			parser := NewParser(0, 1, 2, '\t', false)
			arr, err := parser.Parse(r)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(arr))
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo"}}, Amount: 1.23, Note: "note"}, arr[0])
			assert.Equal(t, ParsedRequests{venmoBatches: [][]string{{"venmo-2"}}, Amount: 1.24, Note: "note-diff"}, arr[1])
		})
	})
}
