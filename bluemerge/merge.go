package bluemerge

import (
	"io"
	"os"
)

// 타수의 파일을 하나의 파일로 이어 붙이는 함수
func MERGE(result string, files ...string) bool {

	// 결과 파일 생성
	d, err := os.Create(result)
	if err != nil {
		return false
	}

	// 인자로 부터 머지할 파일 이름 획득
	for _, file := range files { // range로 가변인자의 모든 값을 꺼냄
		// 파일 오픈
		s, err := os.Open(file)
		if err != nil {
			continue
		}

		// 머지 파일 커멘트 추가
		var comment = "/*merge file " + file + "*/;\r\n"
		d.WriteString(comment)

		// 파일 내용 복사
		if _, err := io.Copy(d, s); err != nil {
			continue
		}

		//줄	바꿈 추가
		d.WriteString("\r\n")
		s.Close()
	}
	d.Close()
	return true
}
