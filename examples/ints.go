package main

import (
    "fmt"
    "github.com/stevedonovan/xflag"
)

func main() {
    flags := xflag.NewFlag()
    ints := flags.IntList("#1","0","set of ints")
    flags.ParseArgs()
    sum := 0
    for _,i := range *ints {
        sum += i
    }
    fmt.Println("sum",sum)
}

