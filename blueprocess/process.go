package blueprocess

import (
	_ "bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	_ "strings"

	ps "github.com/mitchellh/go-ps"
)

//n := []int{1, 2, 3, 4, 5}
//r := sum(n...) // ...를 사용하여 가변인자에 슬라이스를 바로 넘겨줌

// Execute ...
func Execute(name string, output string, args ...string) error {
	var cmd *exec.Cmd

	strArgs := make([]string, len(args))
	var i = 0
	for _, value := range args {
		strArgs[i] = value
		i++
	}

	cmd = exec.Command(name, strArgs...)
	out, _ := cmd.Output()
	//fmt.Printf("execute process %s \r\n", name)

	cmd.Wait()

	if output != "" {
		d, err := os.Create(output)
		if err == nil {
			d.Write(out)
		}
		d.Close()
	}

	return nil
}

// ExecutePipeIn ...
func ExecutePipeIn(name string, input string, args ...string) error {
	var cmd *exec.Cmd

	strArgs := make([]string, len(args))
	var i = 0
	for _, value := range args {
		strArgs[i] = value
		i++
	}
	cmd = exec.Command(name, strArgs...)

	in, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}

	_, err = io.WriteString(in, input)
	if err != nil {
		fmt.Println(err)
		return err
	}
	in.Close()

	//fmt.Printf("execute process %s \r\n", name)
	cmd.Wait()
	return nil
}

// CheckDuplicate ...
func CheckDuplicate(name string) bool {
	pss, err1 := ps.Processes()
	if err1 != nil {
		fmt.Println(err1)
		return false
	}

	for _, process := range pss {
		if name == process.Executable() {
			if os.Getpid() != process.Pid() {
				return true
			}
		}
	}
	return false
}

func FindProcess(name string) int {
	pss, err1 := ps.Processes()
	if err1 != nil {
		fmt.Println(err1)
		return -1
	}

	for _, process := range pss {
		if name == process.Executable() {
			if os.Getpid() != process.Pid() {
				return process.Pid()
			}
		}
	}
	return -1
}

func KillProcess(name string) error {
	pid := FindProcess(name)
	if pid == -1 {
		return errors.New("not find process: " + name)
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	fmt.Printf("kill process %s pid: %d\r\n", name, pid)
	return p.Kill()
}

func KillProcessByPid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	fmt.Printf("kill process pid: %d\r\n", pid)
	return p.Kill()
}
