package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"bytes"
	"regexp"
	"strings"
	"net/http"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

const DEFAULT_CONFIG_FILE = "prom-exporter-aggregator.yml"

// Log Error to Console
func logError(e error, msg ...string) {
	if e != nil {
		fmt.Printf("ERROR: %s\nCaused by: %s\n", msg, e.Error())
	}
}

// Fatal Error handler
func assertNoError(e error, msg string) {
	if e != nil {
		logError(e, msg)
		os.Exit(1)
	}
}

func main() {

	// Parse command line arguments
	var port = flag.String("port", "9191", "Port to listen at")
	var configFile = flag.String("config", DEFAULT_CONFIG_FILE, "Path to config file")
	flag.Parse()
	// Load configuration file
	config := map[string]string{}
	dat, err := ioutil.ReadFile(*configFile)
	assertNoError(err, "Unable to open configuration file. Use --config=file.yml")
	err = yaml.Unmarshal([]byte(dat), &config)
	assertNoError(err, "Invalid configuration file syntax")

	regex := regexp.MustCompile(`(?:(#\s(?:TYPE|HELP))\s)?(\w+)\s(.*)`)

	http.HandleFunc("/metrics", func (w http.ResponseWriter, r *http.Request) {
		var metrics bytes.Buffer
		for url, alias := range config {
			startTime := time.Now()
			fmt.Printf("Querying URL %s ...", url)
			response, err := http.Get(url)
			logError(err)
			if err == nil {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				logError(err)
				if err == nil && (response.StatusCode == 200) {
					for _, line := range strings.Split(string(contents),"\n") {
						all_tokens:= regex.FindStringSubmatch(line)
						if (all_tokens != nil) {
							metric_name := fmt.Sprintf("%s_%s", alias, all_tokens[2])
							wanted_tokens := []string {all_tokens[1], metric_name, all_tokens[3]}
							metrics.WriteString(strings.TrimSpace(strings.Join(wanted_tokens, " ")) + "\n")
						}
					}
				}
			}
			duration := time.Since(startTime)
			fmt.Printf("\tquery took: %s\n", duration)
		}
		fmt.Fprintf(w, metrics.String())
	})
	http.ListenAndServe(":" + *port, nil)
}
