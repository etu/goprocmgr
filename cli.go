package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Cli struct {
	config *Config
}

func (cli *Cli) List(format string) {
	var state ServeFullState

	// Build URL based on config
	requestUrl := fmt.Sprintf("http://%s:%d/api/state", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	// Do request to running instance of program
	res, err := http.Get(requestUrl)

	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	// Validate status code
	if res.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code when fetching active config: %d\n", res.StatusCode)
		os.Exit(2)
	}

	// Read the body content
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(body, &state)

	switch format {
	case "table":
		output := table.NewWriter()
		output.SetOutputMirror(os.Stdout)
		output.AppendHeader(table.Row{"Name", "Running", "Directory", "Command"})

		for _, val := range state.Config.Servers {
			isRunning := false

			if _, ok := state.RunnerState[val.Name]; ok {
				isRunning = true
			}

			output.AppendRow([]interface{}{val.Name, isRunning, val.Directory, val.Command})
		}

		output.Render()

	case "csv":
		output := csv.NewWriter(os.Stdout)
		defer output.Flush()

		output.Write([]string{"Name", "Running", "Directory", "Command"})

		for _, val := range state.Config.Servers {
			isRunning := false

			if _, ok := state.RunnerState[val.Name]; ok {
				isRunning = true
			}

			output.Write([]string{val.Name, fmt.Sprintf("%t", isRunning), val.Directory, val.Command})
		}
	}
}

func (cli *Cli) Add(command string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/config/server", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	directory, err := os.Getwd()

	if err != nil {
		log.Printf("Failed to get current working directory: %s\n", err)
		os.Exit(3)
	}

	// Check if DIRENV_FILE is set or not
	useDirenv := false
	if _, ok := os.LookupEnv("DIRENV_FILE"); ok {
		useDirenv = true
	}

	// Build a new server config
	server := ServerConfig{
		Name:      filepath.Base(directory),
		Command:   command,
		Directory: directory,
		UseDirenv: useDirenv,
		Environment: map[string]string{
			"PATH": os.Getenv("PATH"),
		},
	}

	// Encode the server config as bytes
	body, _ := json.Marshal(server)

	// Pass new buffer for request with URL to post.
	// This will make a post request and will share the JSON data
	res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(body))

	// An error is returned if something goes wrong
	if err != nil {
		panic(err)
	}

	// Need to close the response stream, once response is read.
	// Hence defer close. It will automatically take care of it.
	defer res.Body.Close()

	// Check response code, if New user is created then read response.
	if res.StatusCode == http.StatusCreated {
		log.Println("Created")
	} else {
		var response map[string]string
		resbody, _ := io.ReadAll(res.Body)

		// Parse the json
		json.Unmarshal(resbody, &response)

		// The status is not Created. print the error.
		log.Printf("Failed to create server with response: %s", resbody)
	}
}

func (cli *Cli) Remove(name string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/config/server/%s", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort, name)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, requestUrl, nil)

	if err != nil {
		panic(err)
	}

	// Perform request
	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if res.StatusCode == http.StatusOK {
		log.Println("OK")
		os.Exit(0)
	}

	// Handle error
	var response map[string]string
	resbody, _ := io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(resbody, &response)

	// The status is not Created. print the error.
	log.Printf("Failed to create server with response: %s", resbody)
	os.Exit(4)
}

func (cli *Cli) Start(name string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/runner/%s", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort, name)

	// Pass new buffer for request with URL to post.
	// This will make a post request and will share the JSON data
	res, err := http.Post(requestUrl, "application/json", nil)

	// An error is returned if something goes wrong
	if err != nil {
		panic(err)
	}

	// Check response code, if New user is created then read response.
	if res.StatusCode == http.StatusCreated {
		log.Println("Started")
	} else {
		var response map[string]string
		resbody, _ := io.ReadAll(res.Body)

		// Parse the json
		json.Unmarshal(resbody, &response)

		// The status is not Created. print the error.
		log.Printf("Failed to start server with response: %s", resbody)
	}
}

func (cli *Cli) Stop(name string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/runner/%s", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort, name)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, requestUrl, nil)

	if err != nil {
		panic(err)
	}

	// Perform request
	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if res.StatusCode == http.StatusOK {
		log.Println("OK")
		os.Exit(0)
	}

	// Handle error
	var response map[string]string
	resbody, _ := io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(resbody, &response)

	// The status is not Created. print the error.
	log.Printf("Failed to stop server with response: %s", resbody)
	os.Exit(4)
}

func (cli *Cli) Logs(name string) {
	var state ServeFullState
	stdoutMaxIndex := -1
	stderrMaxIndex := -1

	// Build URL to fetch runners
	runnerRequestUrl := fmt.Sprintf("http://%s:%d/api/state", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	for {
		// Do request to running instance of program to get running processes
		runnerRes, err := http.Get(runnerRequestUrl)

		if err != nil {
			log.Printf("Failed to connect to running instance of program: %s\n", err)
			os.Exit(1)
		}

		// Validate status code
		if runnerRes.StatusCode != http.StatusOK {
			log.Printf("Unexpected status code when fetching runner processes: %d\n", runnerRes.StatusCode)
			os.Exit(2)
		}

		// Read the body content
		defer runnerRes.Body.Close()
		runnerBody, _ := io.ReadAll(runnerRes.Body)

		// Parse the json
		json.Unmarshal(runnerBody, &state)

		if _, ok := state.RunnerState[name]; !ok {
			log.Printf("Process '%s' doesn't seem to be running", name)
			os.Exit(3)
		}

		// Tail stdout
		go func() {
			for key, val := range state.RunnerState[name].Stdout {
				if key > stdoutMaxIndex {
					fmt.Println("stdout>", val)
					stdoutMaxIndex = key
				}
			}
		}()

		// Tail stderr
		go func() {
			for key, val := range state.RunnerState[name].Stderr {
				if key > stderrMaxIndex {
					fmt.Println("stderr>", val)
					stderrMaxIndex = key
				}
			}
		}()

		time.Sleep(1 * time.Second)
	}
}
