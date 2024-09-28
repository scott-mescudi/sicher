package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	srcfile := "srcf/diamond.exe"
	dstdir := "dstf"
	fd := filepath.Join(dstdir, filepath.Base(srcfile))


	ok, err := FileCheck(srcfile, fd)
	if err != nil{
		fmt.Println(err, ok)
		return
	}

	fmt.Println(ok)

	printMemUsage()
}

func FileCheck(srcFile, dstfile string) (bool, error){
	f1, err := os.Stat(srcFile)
	if err != nil{
		return false, fmt.Errorf("error accessing %v: %v", srcFile, err)
	}

	f2, err := os.Stat(dstfile)
	if err != nil{
		return false, fmt.Errorf("error accessing %v: %v", dstfile, err)
	}

	if f1.Size() != f2.Size(){
		return true, nil
	}

	return false,  fmt.Errorf("%v and %v are the same size", srcFile, dstfile)
	
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

