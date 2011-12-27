// it's possible to use xflags to read a simple property file
package main

import (
    "fmt"
    "io/ioutil"
    "github.com/stevedonovan/xflag"
)

var config = `
# comment line is ignored
name=bonzo dog
age=12  # age of animal
owners=Alice,John
`

var (
    flags = xflag.NewFlag()
    name = flags.String("name","dolly","name of object")
    age = flags.Int("age",40,"age of object")
    owners = flags.StringList("owners","self","owners of this animal")
)

func main() {
    tmp,_ := ioutil.TempFile("","tmp")
    tmp.WriteString(config)
    tmp.Close()
    flags.ParseConfig(tmp.Name())
    fmt.Printf("name %q %d owners: %v\n",*name,*age,*owners)
}

