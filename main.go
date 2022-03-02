package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const ShellToUse = "bash"

func parseFlags() string {
	strToSearch := flag.String("s", "", "Use -s stringtosearch")
	flag.Parse()
	return *strToSearch
}

func main() {
	s := parseFlags()
	err, out, errout := Shellout("docker ps -a --format '{{.ID}}'")
	if err != nil {
		log.Fatal(err)
	}
	if len(errout) > 1 {
		fmt.Println("err", errout)
	}
	fmt.Println(errout)
	containerIds := strings.Fields(out)
	for _, v := range containerIds {
		formatedDockerCommand := fmt.Sprintf("docker logs %s", v)
		err, out, errout := Shellout(formatedDockerCommand)
		if err != nil {
			log.Fatal(err)
		}
		contains := strings.Contains(out, s)
		if contains {
			fmt.Println(out)
		}
		if len(errout) > 1 {
			fmt.Println(errout)
		}
	}
}

func Shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}
