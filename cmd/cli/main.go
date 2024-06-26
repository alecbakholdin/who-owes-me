package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"who-owes-me/lib/request_parsing"
	"who-owes-me/lib/venmo"
)

const defaultClientPath = "client.json"

func main() {
	username := flag.String("username", "", "Venmo username")
	password := flag.String("password", "", "Venmo password")
	file := flag.String("f", "", "Optional input file")

	venmoCol := flag.Int("venmo", 0, "Index (0 is first col) of Venmo column in input data")
	amtCol := flag.Int("amount", 1, "Index (0 is first col) of Amount column in input data")
	noteCol := flag.Int("note", 2, "Index (0 is first col) of Note column in input data")
	separator := flag.String("sep", "\t", "Column Separator rune for input data")
	hasHeader := flag.Bool("header", false, "Ignore first line, since it is a header")
	flag.Parse()

	client, err := venmo.LoadClient(defaultClientPath)
	if err != nil {
		fmt.Printf("Error loading client (%s), loading new client\n", err)
		client = venmo.NewClient()
	}
	defer client.StoreToFile(defaultClientPath)

	if *venmoCol == *amtCol || *venmoCol == *noteCol || *amtCol == *noteCol {
		panic(fmt.Sprintf("Cols must be distinct indices, but found venmo[%d], amount[%d], note[%d]", *venmoCol, *amtCol, *noteCol))
	} else if *username == "" || *password == "" {
		panic("Username and password are both required")
	} else if len(*separator) != 1 {
		panic("Separator must be a single character")
	}
	input := getInput(*file)
	defer input.Close()

	parser := request_parsing.NewParser(*venmoCol, *amtCol, *noteCol, rune((*separator)[0]), *hasHeader)
	requests, err := parser.Parse(input)
	if err != nil {
		panic(fmt.Sprintf("error parsing input: %s", err))
	}

	for _, request := range requests {
		for _, batch := range request.VenmoBatches {
			venmoStr := batch[0]
			if len(batch) > 1 {
				venmoStr = fmt.Sprintf("%d venmos", len(batch))
			}
			fmt.Printf("%-30s %6.02f  %s\n", request.Note, request.Amount, venmoStr)
		}
	}
}

func getInput(filepath string) io.ReadCloser {
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			panic(fmt.Sprintf("error opening input file: %s", err))
		}
		return file
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Sprintf("error getting stat for stdin: %s", err))
	} else if stat.Size() == 0 {
		panic("no input provided. Pipe something into stdin or specify a file with -f")
	}
	return os.Stdin
}
