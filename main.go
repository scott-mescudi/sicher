package main

import (
	"fmt"
	"gobackup/src"
	"path/filepath"

	"runtime"
	"os"
)

func main() {
	srcDir := "srcf"
	dstdir := "dstf"
	
	var srcfiles = make(map[string]bool)
	defer src.Clean(srcfiles, dstdir)

	filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path != srcDir{
				srcfiles[path] = true
			}
			
			return nil
	})

	for i := range srcfiles{
		dstfile := filepath.Join(dstdir, i)
		srcfile := filepath.Join(srcDir, i)

		ok, err := src.FileCheck(srcfile, dstfile)
		if err != nil || !ok{
			continue
		}

		src.CopyFile(srcfile, dstfile, 1024)
	}

}






func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

