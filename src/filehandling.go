package src

import (
	"io"
	"os"
	"fmt"
)


func FileCheck(srcFile, dstfile string) (bool, error){
	f1, err := os.Stat(srcFile)
	if err != nil{
		return false, fmt.Errorf("error accessing %v: %v", srcFile, err)
	}

	f2, err := os.Stat(dstfile)
	if err != nil{
		return true, nil
	}

	if f1.Size() != f2.Size(){
		return true, nil
	}

	return false,  fmt.Errorf("%v and %v are the same size", srcFile, dstfile)
	
}

func CopyFile(srcFilePath, dstFilePath string, chunkSize int) (error){
	file, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	
	f, err := os.Create(dstFilePath)
	if err != nil{
		return err
	}
	defer f.Close()

	buf := make([]byte, chunkSize)


	for {
		bytesRead, err  := file.Read(buf)
		if err != nil{
			if err != io.EOF{
				return err
			}
			break
		}

		_, err =  f.Write(buf[:bytesRead])
		if err != nil{
			return err
		}

		if bytesRead < chunkSize {
			break
		}
	}

	return nil
}