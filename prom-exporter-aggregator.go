package main

import (
	"os"
	"fmt"
	"flag"
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
	var configFile = flag.String("config", DEFAULT_CONFIG_FILE, "Path to config file")
	flag.Parse()
	// Load configuration file
	config := map[string]string{}
	dat, err := ioutil.ReadFile(*configFile)
	assertNoError(err, "Unable to open configuration file. Use --config=file.yml")
	err = yaml.Unmarshal([]byte(dat), &config)
	assertNoError(err, "Invalid configuration file syntax")

	// metrics := map[string]string{}

	r := regexp.MustCompile(`(?:(#\s(?:TYPE|HELP))\s)?(\w+)\s(.*)`)

	for url, alias := range config {
		response, err := http.Get(url)
		logError(err)
		if err == nil {
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			logError(err)
			if err == nil && (response.StatusCode == 200) {
				for _, line := range strings.Split(string(contents),"\n") {
					// fmt.Printf("%s\n", line)
					result:= r.FindStringSubmatch(line)
					if (result != nil) {
						metric_name := fmt.Sprintf("%s_%s", alias, result[2])
						reg := []string {result[1], metric_name, result[3]}
						fmt.Printf("%s\n", strings.TrimSpace(strings.Join(reg[:], " ")))
					}
					// for k, v := range result {
					// 	fmt.Printf("%d. %s\n", k, v)
					// }
				}
			}
		}
	}

}
