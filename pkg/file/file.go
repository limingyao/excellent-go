package file

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os"
)

func ReadLines(path string, ch chan<- string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) { //文件已经结束
				break
			}
			log.Println(err)
		}
		ch <- string(bytes.TrimRight(line, "\n"))
	}
	return nil
}

func WriteLines(path string, ch <-chan string) error {
	tmpFilePath := path + "~"
	f, err := os.Create(tmpFilePath)
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(f)

	for line := range ch {
		if _, err := buf.WriteString(line + "\n"); err != nil {
			log.Println(err)
		}
	}

	if err := buf.Flush(); err != nil {
		log.Println(err)
	}
	if err := f.Close(); err != nil {
		log.Println(err)
	}
	if err := os.Rename(tmpFilePath, path); err != nil {
		log.Println(err)
	}

	return nil
}

func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func Copy(src, dst string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	if written, err = io.Copy(dstFile, srcFile); err != nil {
		return 0, err
	}
	if err = dstFile.Sync(); err != nil {
		return 0, err
	}
	return written, nil
}
