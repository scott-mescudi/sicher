package main

import (
	"fmt"
	"gobackup/src"
	"path/filepath"
	"strings"
	"time"

	"os"
)

//TODO 1: add concurrency to clean func
//TODO 2: make each file a go routine with limit of 10 and 100mb per file
//TODO 3: implement worker pool for TODO 2

func main() {
	s := backup{
		"Srcf",
		"dstf",
	}
	start := time.Now()
	s.StartBackup()
	fmt.Printf("Elapsed: %v\n", time.Since(start))
}

var NotAllowedFiles = map[string]bool{
	"test.exe": true,
}

var NotAllowedExt = map[string]bool{
	".sigma": true,
}

func (s backup) StartBackup() {
	var srcfiles = make(map[string]bool)
	defer src.Clean(s.srcDir, s.dstDir)

	filepath.WalkDir(s.srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if _, ok := NotAllowedFiles[path]; ok {
			return filepath.SkipDir
		}

		ext := filepath.Ext(path)
		if _, ok := NotAllowedExt[ext]; ok {
			return filepath.SkipDir
		}

		if path != s.srcDir {
			srcfiles[path] = true
		}

		return nil
	})

	for i := range srcfiles {

		x := strings.TrimPrefix(i, s.srcDir)
		dstfile := filepath.Join(s.dstDir, x)
		srcfile := filepath.Join(i)
		work(srcfile, dstfile, 1024)

	}

}

func work(srcfile, dstfile string, buf int) {
	ok, err := src.FileCheck(srcfile, dstfile)
	if err != nil || !ok {
		return
	}

	err = src.CopyFile(srcfile, dstfile, buf)
	if err != nil {
		fmt.Println(err)
	}
}

type backup struct {
	srcDir string
	dstDir string
}
