xflag extends flag and provides methods to open files for reading and writing.

    import "github.com/stevedonovan/xflag"
    
    var flag = xflag.NewFlag()
    var ofp = flag.CreateFile("o","stdout","output file")
    var iflp = flag.OpenFileList("#1","stdin","input files")
    
Here `ofp` is type `**os.File`, and `iflp` is `*[]string`. This will verify that all 
parameters on the command-line can be opened for reading.

Otherwise, all the methods of package flag are available.

You can read in simple configuration files with `flag.ParseConfig` and have the
lines mapped to flag variables.

Steve Donovan, Copyright 2011
Licence: MIT/X11
