package main

import (
	"fmt"
	"os"
)

func main() {
	os.MkdirAll("testdata", 0755)
	f, err := os.Create("testdata/large_sample.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "name,age,email,city")
	for i := 0; i < 1_000_000; i++ {
		fmt.Fprintf(f, "User%d,%d,user%d@example.com,City%d\n", i, 20+(i%50), i, i%100)
	}
}
