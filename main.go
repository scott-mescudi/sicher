package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"crypto/sha256"
	"io"
	"gobackup/src"
	"os"
)

func main() {
	srcfile := "srcf/test.txt"
	dstdir := "dstf"
	fd := filepath.Join(dstdir, filepath.Base(srcfile))

	ok, err := FileCheck(srcfile, fd)
	if err != nil || !ok{
		fmt.Println(err)
		return
	}
	fmt.Println("copying....")
	src.CopyFile(srcfile, fd, 1000)

	
}


func checksum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil { 
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func FileCheck(srcFile, dstfile string) (bool, error){
	_, err := os.Stat(srcFile)
	if err != nil{
		return false, fmt.Errorf("error accessing %v: %v", srcFile, err)
	}

	_, err = os.Stat(dstfile)
	if err != nil{
		return true, nil
	}

	f1, _ := checksum(srcFile)
	f2, _ := checksum(dstfile)

	if f1 != f2{
		return true, nil
	}
	

	return false, fmt.Errorf("%v and %v are the same size", srcFile, dstfile)
	
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

