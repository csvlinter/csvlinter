package main

import (
    "fmt"
    "os"
)

func main() {
    outPath := "benchmarks/large_invalid_sample.csv"
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
        if i%10 == 0 {
            // Invalid row: empty city, non-numeric age, or invalid email
            fmt.Fprintf(f, "User%d,notanumber,invalid-email,\n", i)
        } else {
            fmt.Fprintf(f, "User%d,%d,user%d@example.com,City%d\n", i, 20+(i%50), i, i%100)
        }
    }
} 