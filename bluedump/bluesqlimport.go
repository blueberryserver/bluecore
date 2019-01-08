package bluedump

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"process"
)

// CreateDatabase ...
func CreateDatabase(host string, port string, user string, pw string, databases []string) error {
	for _, database := range databases {

		log.Printf("create database %s\r\n", database)

		err := process.Execute("mysqladmin.exe", "", "-h", host, "--port", port, "-u", user, pw, "create", database)
		if err != nil {
			return err
		}
	}
	return nil
}

// SQLImport ...
func SQLImport(host string, port string, user string, pw string, filepath string, databases []string, dumpfilenames []string) error {

	dbNames := make([]string, 0)
	for index, database := range databases {
		dumpfile := fmt.Sprintf("%s//%s", filepath, dumpfilenames[index])
		log.Printf("import database:%s <- dump:%s\r\n", database, dumpfilenames[index])

		f, err := os.Open(dumpfile)
		defer f.Close()
		if err != nil {
			log.Printf("%s\r\n", err)
			return err
		}

		dump, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("%s\r\n", err)
			return err
		}
		err = process.ExecutePipeIn("mysql.exe", string(dump), "-h", host, "--port", port, "-u", user, pw, database)
		if err != nil {
			log.Printf("%s\r\n", err)
			return err
		}
		log.Printf("Complete %s\r\n", dumpfile)

		dbNames = append(dbNames, database)
	}

	return nil
}
