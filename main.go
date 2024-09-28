package main

import (
	"fmt"
	"gobackup/src"
	"path/filepath"
	"strings"

	"os"
)

func main() {
	srcDir := "srcf"
	dstdir := "dstf"
	
	var srcfiles = make(map[string]bool)
	defer src.Clean(srcDir, dstdir)

	filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path != srcDir{
				srcfiles[path] = true
			}
			
			return nil
	})

	for i := range srcfiles {
		x := strings.TrimPrefix(i, srcDir)
		
		dstfile := filepath.Join(dstdir, x)
		srcfile := filepath.Join(i)

		ok, err := src.FileCheck(srcfile, dstfile)
		if err != nil || !ok {
			continue
		}

		err = src.CopyFile(srcfile, dstfile, 1024)
		if err != nil{
			fmt.Println(err)
		}
	}


}





