package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func parseFlags() string {
	strToSearch := flag.String("s", "", "Use -s stringtosearch")
	flag.Parse()
	return *strToSearch
}

func main() {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	reader, err := cli.ContainerLogs(context.Background(), "dcbd8eb49e", types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Fatal(err)
	}

	defer reader.Close()

	logContent, err := io.ReadAll(reader)

	if err != nil {
		log.Fatal(err)
	}

	logContentArr := strings.Split(string(logContent), "\n")
	stringToSearch := parseFlags()
	for _, v := range logContentArr {
		contains := strings.Contains(v, stringToSearch)
		if contains {
			fmt.Println(v)
		}
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		log.Fatal(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}
