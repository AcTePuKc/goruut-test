package main

import "os"
import "fmt"
import "github.com/neurlang/classifier/parallel"
import "bufio"
import "strings"
import "math/rand"
import "github.com/neurlang/goruut/pkg/gistselect"

type gistFilterConfig struct {
	enabled bool
	selectK int
	lambda  float64
}

func load(filename string, top int, gistFilter gistFilterConfig) [][2]string {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	entries := make([]gistselect.Entry, 0, 256)

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")

		// Check if we have exactly two or three columns
		if len(columns) != 2 && len(columns) != 3 {
			fmt.Println("Line does not have exactly two or three columns:", line)
			continue
		}

		// Process each column
		entry := gistselect.Entry{
			Word: columns[0],
			IPA:  columns[1],
		}
		if len(columns) == 3 {
			entry.Extra = columns[2]
		}
		entries = append(entries, entry)
	}

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	if gistFilter.enabled {
		selectK := gistFilter.selectK
		if selectK < 0 {
			selectK = 0
		}
		cfg := gistselect.Config{
			K:        selectK,
			Lambda:   gistFilter.lambda,
			Distance: gistselect.JointLevenshtein,
			Utility:  gistselect.CoverageUtility{},
		}
		result := gistselect.Select(entries, cfg)
		entries = result.Entries
	}

	slice := make([][2]string, 0, len(entries))
	for _, entry := range entries {
		slice = append(slice, [2]string{entry.Word, entry.IPA})
	}
	if top >= 0 {
		rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
		if len(slice) > top {
			slice = slice[:top]
		}
	}
	return slice
}

func loop(slice [][2]string, group int, do func(string, string)) {
	parallel.ForEach(len(slice), group, func(n int) {
		// Process each column
		column1 := slice[n][0]
		column2 := slice[n][1]
		// Example: Print the columns
		do(column1, column2)
	})
}
