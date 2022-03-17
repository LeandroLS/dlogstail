package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Container struct {
	Name  string
	Id    string
	Image string
}

type Containers []Container

type Logs struct {
	Content    string
	LineByLine []string
}

var (
	//go:embed home.html
	htmlFile embed.FS
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(htmlFile, "*.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}

func containersHandler(w http.ResponseWriter, r *http.Request) {
	var containers Containers
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	containersRaw, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, container := range containersRaw {
		containers = append(containers, Container{Name: container.Names[0], Id: container.ID[:15], Image: container.Image})
	}
	json.NewEncoder(w).Encode(containers)
}

func tailLog(logs []string, newLogsLenght int) []string {
	diff := len(logs) - newLogsLenght
	return logs[diff-1 : newLogsLenght+diff]
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	containerId := r.URL.Query().Get("container_id")
	numberOfLinesS := r.URL.Query().Get("number_of_lines")
	//find a way to delete this var
	var numberOfLinesI int
	if numberOfLinesS != "" {
		numberOfLinesI, _ = strconv.Atoi(numberOfLinesS)
	}
	//todo remover esse cli
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	reader, err := cli.ContainerLogs(context.Background(), containerId, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	logContent, err := io.ReadAll(reader)
	logsSplitedByNewLine := strings.Split(string(logContent), "\n")
	if numberOfLinesI <= len(logsSplitedByNewLine) {
		logsSplitedByNewLine = tailLog(logsSplitedByNewLine, numberOfLinesI)
	}
	for i, v := range logsSplitedByNewLine {
		logsSplitedByNewLine[i] = strings.TrimSpace(v)
	}
	if err != nil {
		log.Fatal(err)
	}
	logs := Logs{Content: string(logContent), LineByLine: logsSplitedByNewLine}
	json.NewEncoder(w).Encode(logs)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/containers", containersHandler)
	http.HandleFunc("/containers/logs", logsHandler)
	fmt.Println("dlogstail is running")
	http.ListenAndServe(":3001", nil)
}
