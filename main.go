package main

import (
	"fmt"

	"os"
	"path/filepath"
	"runtime"

	"gobackup/src"
)

func main() {
	srcfile := "srcf/diamond.exe"
	dstdir := "dstf"

	if fn, err := FileCheck(srcfile, dstdir); err != nil{
		src.CopyFile(srcfile, fn,  104_857_600)
		return
	}

	fmt.Println("file in dst dir")
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

