package main

import (
    "fmt"
    "os"
    "github.com/stevedonovan/xflag"
)

func main() {
    flags := xflag.NewFlag()
    fl := flags.OpenFileList("#1","","input files")
    flags.ParseArgs()
    slice := make([]byte,20)
    for _,fname := range *fl {
        fmt.Println("file",fname)
        f,_ := os.Open(fname)
        f.Read(slice)
        fmt.Println(string(slice))
        f.Close()
    }
}

