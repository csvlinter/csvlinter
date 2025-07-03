package main

import (
	"fmt"
	"os"
)

func main() {
	outPath := "benchmarks/large_valid_sample.csv"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}
	os.MkdirAll("benchmarks", 0755)
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "name,age,email,city")
	for i := 0; i < 1_000_000; i++ {
		fmt.Fprintf(f, "User%d,%d,user%d@example.com,City%d\n", i, 20+(i%50), i, i%100)
	}
}
