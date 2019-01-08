package bluecore

import (
	"io"
	"log"
	"os"
)

// InitLog ...
func InitLog(filename string) (*os.File, error) {
	fileLog, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	mutiWriter := io.MultiWriter(fileLog, os.Stdout)
	log.SetOutput(mutiWriter)
	return fileLog, nil
}
