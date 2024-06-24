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
	client, err := venmo.LoadClient(defaultClientPath)
	if err != nil {
		fmt.Printf("Error loading client, loading new client: %s", err)
		client = venmo.NewClient()
	}
	defer client.StoreToFile(defaultClientPath)

	username := flag.String("username", "", "Venmo username")
	password := flag.String("password", "", "Venmo password")
	file := flag.String("f", "", "Optional input file")

	venmoCol := flag.Int("venmo", 0, "Index (0 is first col) of Venmo column in input data")
	amtCol := flag.Int("amount", 1, "Index (0 is first col) of Amount column in input data")
	noteCol := flag.Int("note", 2, "Index (0 is first col) of Note column in input data")
	separator := flag.String("sep", "\t", "Column Separator for input data")
	hasHeader := flag.Bool("header", false, "Ignore first line, since it is a header")
	flag.Parse()

	if venmoCol == amtCol || venmoCol == noteCol || amtCol == noteCol {
		panic(fmt.Sprintf("Cols must be distinct indices, but found venmo[%d], amount[%d], note[%d]", *venmoCol, *amtCol, *noteCol))
	} else if *username == "" || *password == "" {
		panic("Username and password are both required")
	}
	_ = request_parsing.NewParser(*venmoCol, *amtCol, *noteCol, *separator, *hasHeader)
	
	input := getInput(*file)
	defer input.Close()
}

func getInput(filepath string) io.ReadCloser {
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			panic(fmt.Sprintf("error opening input file: %s", err))
		}
		return file
	}

	fmt.Println("Paste your input, then press enter:")
	return os.Stdin
}