package main

import (
   "fmt"
   "github.com/stevedonovan/xflag"
)

func main() {
   flag := xflag.NewFlag()
   f := flag.OpenFile("#1","stdin","input file")
   flag.ParseArgs()
   fi,e := f.Stat()
   if e != nil {
      fmt.Println("error!",e)
   } else {
      fmt.Println(fi.IsDirectory(),fi.IsRegular())
   }
   flag.Close()
}
