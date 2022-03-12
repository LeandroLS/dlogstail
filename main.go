package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func parseFlags() string {
	strToSearch := flag.String("s", "", "Use -s stringtosearch")
	flag.Parse()
	return *strToSearch
}

func getLogs() {
	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	stringToSearch := parseFlags()
	var oldLogContent string

	for true {
		reader, err := cli.ContainerLogs(context.Background(), "dcbd8eb49e", types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})

		if err != nil {
			log.Fatal(err)
		}

		defer reader.Close()

		logContent, err := io.ReadAll(reader)

		if err != nil {
			log.Fatal(err)
		}

		logContentIsSame := strings.Compare(oldLogContent, string(logContent))

		if logContentIsSame == -1 {
			logContentArr := strings.Split(string(logContent), "\n")

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

			oldLogContent = string(logContent)
		}

	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("home.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}

type Container struct {
	Name  string
	Id    string
	Image string
}

type Containers []Container

type Logs struct {
	Content string
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

		containers = append(containers, Container{Name: container.Names[0], Id: container.ID, Image: container.Image})
	}

	json.NewEncoder(w).Encode(containers)
}

//todo get container information dynamically
func logsHandler(w http.ResponseWriter, r *http.Request) {

	queryValues := r.URL.Query()

	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		log.Fatal(err)
	}

	reader, err := cli.ContainerLogs(context.Background(), queryValues["container_id"][0], types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})

	if err != nil {
		log.Fatal(err)
	}

	defer reader.Close()

	logContent, err := io.ReadAll(reader)

	if err != nil {
		log.Fatal(err)
	}

	logs := Logs{Content: string(logContent)}

	json.NewEncoder(w).Encode(logs)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/containers", containersHandler)
	http.HandleFunc("/logs", logsHandler)
	http.ListenAndServe(":3001", nil)
}
