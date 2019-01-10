package bluecp

import (
	"io"
	"os"
)

// 파일 복사
func CP(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

// 파일, 디렉토리 삭제
func RM(name string) error {

	err := os.Remove(name)
	if err != nil {
		return err
	}
	return nil
}

// 디렉토리 생성
func MKDIR(name string, permission uint32) error {

	err := os.Mkdir(name, os.FileMode(permission))
	if err != nil {
		return err
	}
	return nil
}

// 작업 경로 이동
func CD(dir string) error {
	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	return nil
}
