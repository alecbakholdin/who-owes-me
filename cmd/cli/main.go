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

	totalSent := float64(0)
	totalRequested := float64(0)
	for _, request := range requests {
		for _, batch := range request.VenmoBatches {
			venmoStr := batch[0]
			if len(batch) > 1 {
				venmoStr = fmt.Sprintf("%d venmos", len(batch))
			}
			fmt.Printf("%-30s %6.02f  %s\n", request.Note, request.Amount, venmoStr)
			if request.Amount > 0 {
				totalRequested += request.Amount * float64(len(batch))
			} else {
				totalSent += -1 * request.Amount * float64(len(batch))
			}
		}
	}
	width := 25
	fmt.Printf("%-*s %0.02f\n", width, "Total Being Sent:", totalSent)
	fmt.Printf("%-*s %0.02f\n", width, "Total Requested:", totalRequested)
	fmt.Printf("%-*s %0.02f\n", width, "Total Net:", totalRequested - totalSent)
	fmt.Printf("Confirm these transactions? (y/n)")
	var str string
	if n, err := fmt.Scan(&str); err != nil {
		panic(fmt.Sprintf("error reading from stdin: %s", err))
	} else if n != 1 {
		panic(fmt.Sprintf("expected one character from stdin, but read %d", n))
	}
	r  := str[0]
	if r == 'y' || r == 'Y' || r == '\n' || r == '\r'  {
		fmt.Println("Sending requests")
	} else if r == 'n' || r == 'N' {
		return
	} else {
		panic(fmt.Sprintf("unexpected character %c", r))
	}

	client, err := venmo.LoadClient(defaultClientPath)
	if err != nil {
		fmt.Printf("Error loading client (%s), loading new client\n", err)
		client = venmo.NewClient()
	}
	defer client.StoreToFile(defaultClientPath)
	if err := client.Login(*username, *password); err != nil {
		panic(fmt.Sprintf("error logging in: %s", err))
	}
	for _, r := range requests {
		for _, b := range r.VenmoBatches {
			if err := client.ProcessPayment(r.Amount, r.Note, b); err != nil {
				fmt.Printf("Error processing batch %s %.02f %v: %s\n", r.Note, r.Amount, b, err)
			}
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
