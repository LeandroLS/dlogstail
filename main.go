package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	w.Write([]byte("<h1>Hello World!</h1>"))
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	h := HomePage{Teste: "teste", ContainerNames: []string{"container1", "container2"}}
	t, err := template.ParseFiles("home.html")
	fmt.Println(err)
	t.Execute(w, h)
}

type HomePage struct {
	ContainerNames []string
	Teste          string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/home", homePageHandler)
	http.ListenAndServe(":"+port, nil)
}
