package src

import (
	"io"
	"os"
	"path/filepath"
)

func FileCheck(srcFile, dstDIR string) (string, error){
	f := filepath.Base(srcFile)
	fn := filepath.Join(dstDIR, f)
	if _, err := os.Stat(fn); err != nil{
		return fn, err
	}

	return "", nil
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