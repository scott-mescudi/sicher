package main

import (
	"fmt"
	"gobackup/src"
	"path/filepath"
	"strings"

	"os"
)

func(s backup) StartBackup(){
	var srcfiles = make(map[string]bool)
	defer src.Clean(s.srcDir, s.dstDir)

	filepath.WalkDir(s.srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path != s.srcDir{
				srcfiles[path] = true
			}
			
			return nil
	})

	for i := range srcfiles {
		x := strings.TrimPrefix(i, s.srcDir)
		
		dstfile := filepath.Join(s.dstDir, x)
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

type backup struct{
	srcDir string
	dstDir string
}

func main() {
	s := backup{
		"Srcf",
		"dstf",
	}

	s.StartBackup()
}





