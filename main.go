package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gobackup/src"
)

func main() {
	srcfile := "srcf/diamond.exe"
	dstdir := "dstf"
	fd := filepath.Join(dstdir, filepath.Base(srcfile))


	ok, err := src.FileCheck(srcfile, fd)
	if err != nil || !ok{
		fmt.Println(err)
		return
	}
	fmt.Println("copying....")
	src.CopyFile(srcfile, fd, 1000)

	printMemUsage()
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

