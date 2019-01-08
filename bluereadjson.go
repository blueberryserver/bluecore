package bluecore

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ex)
// type Config struct {
// 	Host string `json:"host"`
// 	User string `json:"user"`
// 	Pw   string `json:"pw"`
// 	Line string `json:"line"`
// 	Path string `json:"path"`
// }
// var config = &Config{}
// err := bluecore.ReadJSON(config, "conf.json")

// ReadJSON json 파일 읽기
func ReadJSON(v interface{}, filename string) error {
	isUtf8, err := CheckUtf8(filename)
	if err != nil {
		fmt.Println(err)

		// change file string encoding -> utf8
		if isUtf8 == false {
			temp := strings.Split(filename, ".")
			var changedfilename = fmt.Sprintf("%s_utf8.%s", temp[0], temp[1])
			if err := EncodingfileUtf8(filename, changedfilename); err != nil {
				return err
			}
			filename = changedfilename
		} else {
			return err
		}
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	//var input inputData
	jsonParser := json.NewDecoder(file)
	if err := jsonParser.Decode(v); err != nil {
		return nil
	}
	return nil
}
