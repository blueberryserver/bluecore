package bluecore

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// ReadYAML ...
func ReadYAML(v interface{}, filename string) error {
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

	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = yaml.Unmarshal(fileData, v)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// WriteYAML ...
func WriteYAML(v interface{}, filename string) error {
	result, err := yaml.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return err
	}

	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(result)
	if err != nil {
		return err
	}

	return nil
}
