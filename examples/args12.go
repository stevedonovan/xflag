package main

import (
    "fmt"
    "github.com/stevedonovan/xflag"
)

func main() {
    flags := xflag.NewFlag()
    name := flags.String("#1","dolly","name of object")
    age := flags.Int("#2",40,"age of object")
    flags.ParseArgs()
    fmt.Println("name",*name,"age",*age)
}

