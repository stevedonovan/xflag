// Copyright, 2011 Steve Donovan
// License MIT/X11

/*  
    Package xflag extends flag and provides methods to open files for reading and writing.
  
    Usage:
  
    It extends the flag.FlagSet type, so a FlagExtra must be explicitly
    created with xflag.NewFlag().
        import "xflag"
        var flag = xflag.NewFlag()
        var ip = flag.Int("d",404,"increment delta")
        var fp = flag.OpenFile("i","stdin","input file")
    OpenFile and CreateFile return pointers to the value, like the rest of
    the flag functions - in this case **os.File.
    You may specify non-flag parameters with #1, #2, etc:
        var sp = flag.String("#1","","device name")
    Together with this feature comes varargs support. In this case
    there can be an arbitrary number of integers starting at parameter 2:
        var ilp = flag.IntList("#2","","port list")
    Although typically used for multiple non-flag parameters,
    these both work as expected:
        $ prog1 10 20 30
        $ prog2 -list 10,20,30
    since internally the value is a comma-separated list.
  
    Together with OpenFileList for variable number of input files, comes
    support for file pattern globbing on Windows, where the default
    shell does not do the usual expansion. Note that this function returns
    *[]string, since generally you don't want to open a potentially large
    number of files at once. However, it does guarantee that all of these
    files can be safely opened.
  
    The ParseConfig() method provides a simple way to use flag-style
    variables with configuration files. These are assumed to be <var>=<value>
    pairs with # comments; blank lines are ignored.  Any detected wildcards
    are expanded for OpenFile on all platforms.  Vararg variables use comma
    separated values as above:
        ports=10,20,30  #  pports := flag.IntList("ports"...
        files=*.go,makfile # pfiles := flag.OpenFileList("files"...
*/
package xflag

import (
    "os"
    "strings"
    "strconv"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "runtime"
    "flag"
)


type ListValue interface {
    flag.Value
    ListValue()
}

type FlagExtra struct {
    *flag.FlagSet
    Files []*fileValue
}

func (fx *FlagExtra) quitf(format string, values... interface{}) {
    fmt.Fprintf(os.Stderr,format,values...)
    fmt.Fprintln(os.Stderr,"\n")
    fx.PrintDefaults()
    os.Exit(1)
}

// Create a new FlagExtra, which has all the existing methods of flag.FlagSet,
// plus methods for handling files more transparently.
func NewFlag () *FlagExtra {
    return &FlagExtra{flag.NewFlagSet(os.Args[0],flag.ExitOnError),[]*fileValue{}}
}

func (fx *FlagExtra) addFile (f *fileValue) {
    fx.Files = append(fx.Files,f)
}

// If you have (possibly) opened files with OpenFile or CreateFile,
// then calling this ensures that the files are properly closed, if needed.
func (fx *FlagExtra) Close () {
    for _, f := range fx.Files {
        f.Close()
    }
}

func glob (cmdline []string) []string {
    xcmds := []string{}
    for _,a := range cmdline {
        if strings.IndexAny(a,"*?") > -1 {
            glob,_ := filepath.Glob(a)
            xcmds = append(xcmds,glob...)
        } else {
            xcmds = append(xcmds,a)
        }
    }
    return xcmds
}

// Parse a set of parameters, optionally doing file glob expansion.
func (fx *FlagExtra) Parse(cmdline []string, doGlob bool) {
    if doGlob {
        cmdline = glob(cmdline)
    }
    fx.FlagSet.Parse(cmdline)
    args := fx.Args()
    fx.VisitAll(func (flag *flag.Flag) {
        if flag.Name[0] == '#' {
            idx,e := strconv.Atoi(flag.Name[1:])
            if e != nil {
                fx.quitf("bad flag index " + e.String())
            }
            idx --
            var val string
            if idx < len(args) {
                if _,ok := flag.Value.(ListValue); ok {
                    val = strings.Join(args[idx:],",")
                } else {
                    val = args[idx]
                }

            } else {
                val = flag.DefValue
            }
            res := flag.Value.Set(val)
            if ! res {
                fx.quitf("invalid value %q for argument %d\n",val,idx+1)
            }
        }
    })

}

var isWindows = runtime.GOOS == "windows"

// Use this instead of flag.Parse();
// It allows typed positional arguments that use #n.
func (fx *FlagExtra) ParseArgs() {
    fx.Parse(os.Args[1:],isWindows)
}

// Program arguments may be read from a named configuration file. This
// is similar to the command-line format, except that initial hyphen
// is not used and lines may end with a # comment.
func (fx *FlagExtra) ParseConfig(filename string) {
    bytes, e := ioutil.ReadFile(filename)
    if e != nil {
        fx.quitf("cannot open config file %q",filename)
    }
    // necessary Windows hack
    contents := strings.Replace(string(bytes),"\r\n","\n",-1)
    lines := strings.Split(contents,"\n")
    out := []string{}
    for _,line := range lines {
        //  # comments are removed; empty lines are ignored
        idx := strings.Index(line,"#")
        if idx > -1 {
            line = line[0:idx]
        }
        line = strings.TrimSpace(line)
        if len(line) > 0 {
            // make var=value lines look like command flags
            if strings.Contains(line,"=") {
                line = "-"+line
            }
            out = append(out,line)
        }
    }
    fx.Parse(out,true)
}

type fileValue struct {
    f *os.File
    name string
    in, std bool
}

// implement flag.Value interface
func (this *fileValue) String() string {
    return this.name
}

func (this *fileValue) Set(name string) bool {
    var f *os.File
    var e os.Error
    if len(name) == 0 {
        return false
    } else if name == "stdin" {
        f = os.Stdin
    } else if name == "stdout" {
        f = os.Stdout
    } else {
        this.std = false
        if this.in {
            f,e = os.Open(name)
        } else {
            f,e = os.Create(name)
        }
        if e != nil {
            return false
        }
    }
    this.name = name
    this.f = f
    return true
}

func (this *fileValue) Close() {
    if ! this.std {
        this.f.Close()
    }
}

func newFileValue (name string, in bool) *fileValue {
    return &fileValue{nil,name,in,true}
}

func (fx *FlagExtra) newFileValue(name string, in bool) *fileValue {
    file := newFileValue(name,in)
    fx.addFile(file)
    return file
}

// Opens a file for reading. The default value may be empty (meaning that
// this is a required parameter), a valid file, or "stdin" meaning open
// standard input.
func (fx *FlagExtra) OpenFile(name, def, usage string) **os.File {
    file := fx.newFileValue(def,true)
    fx.Var(file,name,usage)
    return &file.f
}

// Opens a file for writing. An empty default string means a required file
// parameter, and "stdout" means open standard output.
func (fx *FlagExtra) CreateFile(name, def, usage string) **os.File {
    file := fx.newFileValue(def,false)
    fx.Var(file,name,usage)
    return &file.f
}

// a convenient interface for implementing value lists
type Converter interface {
    Convert(value string) bool
}

// a useful base class for value lists; implements ListValue
type ValueList struct {
    name string
    converter Converter
}

func (this *ValueList) String() string {
    return this.name
}

func (this *ValueList) Set(value string) bool {
    values := strings.Split(value,",")
    for _, v := range values {
        if ! this.converter.Convert(v) {
            return false
        }
    }
    return true
}

// mark as satisfying the ListValue interface
func (this *ValueList)  ListValue() {}

func makeValueList(def string) ValueList {
    return ValueList{def,nil}
}

type filesValue struct {
    ValueList
    names []string
}

func (this *filesValue) Convert(v string) bool {
    file := newFileValue(v,true)
    if ! file.Set(v) {
        return false
    }
    // we just make sure we can open this file, so close the handle
    file.Close()
    this.names = append(this.names,file.name)
    return true
}

// Bind a variable number of parameters to a list of files.
// Generally only used with #n names, but the variable parameter functions
// will also work with configuration files
func (fx *FlagExtra) OpenFileList(name, def, usage string) *[]string {
    filelist := &filesValue{makeValueList(def),[]string{}}
    filelist.converter = filelist
    fx.Var(filelist,name,usage)
    return &filelist.names
}

type intsValue struct {
    ValueList
    ints []int
}

func (this *intsValue) Convert(v string) bool {
    i,e := strconv.Atoi(v)
    if e != nil {
        return false
    }
    this.ints = append(this.ints,i)
    return true
}

// Bind a variable number of parameters to a list of integers.
func (fx *FlagExtra) IntList(name, def, usage string) *[]int {
    intlist := &intsValue{makeValueList(def),[]int{}}
    intlist.converter = intlist
    fx.Var(intlist,name,usage)
    return &intlist.ints
}

type stringsValue struct {
    ValueList
    strings []string
}

func (this *stringsValue) Convert(v string) bool {
    this.strings = append(this.strings,v)
    return true
}

// Bind a number of values as a list of strings.
func (fx *FlagExtra) StringList(name, def, usage string) *[]string {
    slist := &stringsValue{makeValueList(def),[]string{}}
    slist.converter = slist
    fx.Var(slist,name,usage)
    return &slist.strings
}

