package main

import (
	"bytes"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const DEFAULT_CONFIG_FILE = "prom-exporter-aggregator.yml"

// Fatal Error handler
func assertNoError(e error, msg string) {
	if e != nil {
		fmt.Printf("ERROR: %s. Cause: %s", msg, e.Error())
		os.Exit(1)
	}
}

func main() {

	// Parse command line arguments
	var port = flag.String("port", "9191", "Port to listen on")
	var configFile = flag.String("config", DEFAULT_CONFIG_FILE, "Path to config file")
	flag.Parse()
	// Load configuration file
	config := map[string]string{}
	dat, err := ioutil.ReadFile(*configFile)
	assertNoError(err, "Unable to open configuration file. Use --config=file.yml")
	err = yaml.Unmarshal([]byte(dat), &config)
	assertNoError(err, "Invalid configuration file syntax")

	var wg sync.WaitGroup
	regex := regexp.MustCompile(`(?:(#\s(?:TYPE|HELP))\s)?(\w+)\s(.*)`)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var metrics bytes.Buffer
		wg.Add(len(config))
		for url, alias := range config {
			go func(url string, alias string) {
				defer wg.Done()
				reqStartTime := time.Now()
				response, err := http.Get(url)
				if err != nil {
					fmt.Printf("ERROR: Endpoint %s. Cause: %s", url, err.Error())
				} else {
					defer response.Body.Close()
					contents, err := ioutil.ReadAll(response.Body)
					if err != nil || response.StatusCode != 200 {
						fmt.Printf("ERROR: Invalid response from %s\n", url)
					} else {
						for _, line := range strings.Split(string(contents), "\n") {
							all_tokens := regex.FindStringSubmatch(line)
							if all_tokens != nil {
								metric_name := fmt.Sprintf("%s_%s", alias, all_tokens[2])
								wanted_tokens := []string{all_tokens[1], metric_name, all_tokens[3]}
								metrics.WriteString(strings.TrimSpace(strings.Join(wanted_tokens, " ")) + "\n")
							}
						}
						fmt.Printf("INFO: Query to endpoint %s took %s\n", url, time.Since(reqStartTime))
					}
				}
			}(url, alias)
		}
		wg.Wait()
		fmt.Fprintf(w, metrics.String())
	})
	http.ListenAndServe(":"+*port, nil)
}
