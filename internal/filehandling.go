package pkg

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

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

func FileCheck(srcFile, dstfile string, sizeLimit int) (bool, error) {
	f , err := os.Stat(srcFile)
	if err != nil {
		return false, fmt.Errorf("error accessing %v: %v", srcFile, err)
	}

	if f.Size() > int64(sizeLimit){
		return false, fmt.Errorf("%v is larger than the specified size limit", srcFile)
	}

	_, err = os.Stat(dstfile)
	if err != nil {
		return true, nil
	}

	f1, _ := checksum(srcFile)
	f2, _ := checksum(dstfile)

	if f1 != f2 || f1 == "" && f2 == "" {
		return true, nil
	}

	return false, fmt.Errorf("%v and %v are the same size", srcFile, dstfile)

}

func CopyFile(srcFilePath, dstFilePath string, chunkSize int) error {
	file, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fs, err := os.Stat(srcFilePath)
	if err != nil {
		return err
	}

	if fs.IsDir() {
		os.Mkdir(dstFilePath, 0700)
		return nil
	}

	f, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, chunkSize)

	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		_, err = f.Write(buf[:bytesRead])
		if err != nil {
			return err
		}

		if bytesRead < chunkSize {
			break
		}
	}

	return nil
}
