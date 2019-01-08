package bluedump

import (
	"fmt"
	"log"
	"process"
)

// SQLExport ...
func SQLExport(host string, port string, user string, pw string, filepath string, databases []string, dumpfilenames []string, options []string) error {

	for index, database := range databases {
		log.Printf("export database:%s -> dump file:%s\r\n", database, dumpfilenames[index])

		dumpfile := fmt.Sprintf("%s//%s", filepath, dumpfilenames[index])

		option := []string{
			"--skip-opt", "--net_buffer_length=409600", "--create-options", "--disable-keys",
			"--routines", "--triggers", "--lock-tables", "--quick", "--set-charset", "--extended-insert", "--single-transaction", "--add-drop-table",
			"--no-create-db"}

		for _, op := range options {
			option = append(option, op)
		}

		{
			option = append(option, "-h")
			option = append(option, host)
			option = append(option, "--port")
			option = append(option, port)
			option = append(option, "-u")
			option = append(option, user)
			option = append(option, pw)
			option = append(option, database)
		}
		log.Printf("%s\r\n", option)

		err := process.Execute("mysqldump.exe", dumpfile, option...)

		if err != nil {
			return err
		}
	}
	return nil
}
