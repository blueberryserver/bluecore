package bluecore

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding/korean"
)

// CheckUtf8 UTF8 파일 스트링 확인
func CheckUtf8(filename string) (bool, error) {
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		fmt.Println("file open errer")
		return true, err
	}

	//var EOF = errors.New("EOF")
	reader := bufio.NewReader(f)
	for {
		r, err := reader.ReadString('\n')
		if false == utf8.ValidString(r) {
			fmt.Printf("invalid utf8 file %s \r\n", r)
			return false, errors.New("invalid utf8 encoding file")
		}

		if err == io.EOF {
			break
		}
	}
	return true, nil
}

// EncodingfileUtf8 파일 UTF8로 인코딩
func EncodingfileUtf8(filename string, resultfilename string) error {
	d, err := os.Create(resultfilename)
	defer d.Close()
	if err != nil {
		return err
	}

	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		fmt.Println("file open errer")
		return err
	}

	r := korean.EUCKR.NewDecoder().Reader(f)
	io.Copy(d, r)
	return nil
}

// EncodingfileAnsi 파일 Ansi로 인코딩(korea.EUCKR)
func EncodingfileAnsi(filename string, resultfilename string) error {
	d, err := os.Create(resultfilename)
	defer d.Close()
	if err != nil {
		return err
	}

	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		fmt.Println("file open errer")
		return err
	}

	r := korean.EUCKR.NewEncoder().Writer(d)
	io.Copy(r, f)
	return nil
}
