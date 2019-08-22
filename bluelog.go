package bluecore

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/blueberryserver/bluecore/bluecp"
)

// InitLog ...
func InitLog(filename string) (*os.File, error) {
	fileLog, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		filePathNameArry := strings.Split(filename, "/")
		path := ""
		for i := 0; i < len(filePathNameArry)-1; {
			fmt.Println(filePathNameArry[i])
			path += filePathNameArry[i]
			i++
		}
		err := bluecp.MKDIR(path, 0755)
		if err != nil {
			return nil, err
		}

		fileLog, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
	}
	mutiWriter := io.MultiWriter(fileLog, os.Stdout)
	log.SetOutput(mutiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return fileLog, nil
}
