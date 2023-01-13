package file

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func ReadFile(path string, ch chan<- string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.WithError(err).Error()
		}
	}()

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if len(line) < 1 && err != nil {
			if errors.Is(err, io.EOF) { //文件已经结束
				break
			}
			log.WithError(err).Error()
		}
		ch <- string(line)
	}
	return nil
}

func WriteFile(path string, ch <-chan string) error {
	tmpFile := path + "~"
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(f)

	for line := range ch {
		if _, err := buf.WriteString(line + "\n"); err != nil {
			log.WithError(err).Error()
		}
	}

	if err := buf.Flush(); err != nil {
		log.WithError(err).Error()
	}
	if err := f.Close(); err != nil {
		log.WithError(err).Error()
	}
	if err := os.Rename(tmpFile, path); err != nil {
		log.WithError(err).Error()
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
