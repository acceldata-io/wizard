package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/acceldata-io/wizard/pkg/wlog"
	"github.com/acceldata-io/wizard/task"
)

const configFile = "./files/example.json"

type templateConfig struct {
	CPUShares   string
	MemoryLimit string
}

func main() {
	fmt.Println("INFO: Running example program")

	// Read the defined JSON file with wizard DSL
	config, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("ERROR: Cannot read the %s file, Reason: %s\n", configFile, err.Error())
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		if os.Args[1] == "chan" {
			withLogsChan(config)
		} else if os.Args[1] == "non-chan" {
			withoutLogsChan(config)
		} else {
			fmt.Println("ERROR: invalid argument. expecting chan or non-chan, got: ", os.Args[1])
			os.Exit(1)
		}
	} else if len(os.Args) == 1 {
		withoutLogsChan(config)
	} else {
		fmt.Println("ERROR: invalid arguments. takes 1 got ", len(os.Args))
		os.Exit(1)
	}
}

func withoutLogsChan(config []byte) {
	// load the tasks, facts and validate the input JSON config
	tasks, err := task.New(config, embed.FS{}, task.TemplateOptions{
		EnableWizardFacts: false,
		TemplateConfig: templateConfig{
			CPUShares:   "1024",
			MemoryLimit: "1G",
		},
	})
	if err != nil {
		fmt.Printf("ERROR: Cannot create tasks, Because: %s\n", err.Error())
		os.Exit(1)
	}

	// Run the tasks and get []interface{} of logs and error if any
	// Ignore the logs if not needed using _
	wLog, err := tasks.Execute()
	if err != nil {
		fmt.Printf("ERROR: Cannot perform task, Because: %s\n", err.Error())
		fmt.Println("ERROR: Stopping all process and exiting")
		os.Exit(1)
	}

	for _, v := range wLog {
		switch v.(type) {
		case wlog.WLInfo:
			fmt.Println("INFO:", v)
		case wlog.WLError:
			fmt.Println("ERROR:", v)
		case wlog.WLWarn:
			fmt.Println("WARN:", v)
		case wlog.WLDebug:
			fmt.Println("DEBUG:", v)
		default:
			fmt.Println("ERROR:", v)
		}
	}
	fmt.Println("INFO: Performed all tasks successfully")
}

func withLogsChan(config []byte) {
	// load the tasks, facts and validate the input JSON config
	tasks, wLog, err := task.NewWithLog(config, embed.FS{}, task.TemplateOptions{
		EnableWizardFacts: false,
		TemplateConfig: templateConfig{
			CPUShares:   "1024",
			MemoryLimit: "1G",
		},
	})
	if err != nil {
		fmt.Printf("ERROR: Cannot create tasks, Because: %s\n", err.Error())
		os.Exit(1)
	}

	// Run the tasks in a go routine and get error if any
	// the wLog created from the NewWithLog should be passed to the Perform()
	// The wLog will push to the chan
	go func() {
		err = tasks.Perform(wLog)
	}()

	// loop through the chan to make it non-blocking
	for v := range wLog {
		switch v.(type) {
		case wlog.WLInfo:
			fmt.Println("INFO:", v)
		case wlog.WLError:
			fmt.Println("ERROR:", v)
		case wlog.WLWarn:
			fmt.Println("WARN:", v)
		case wlog.WLDebug:
			fmt.Println("DEBUG:", v)
		default:
			fmt.Println("ERROR:", v)
		}
	}
	if err != nil {
		fmt.Printf("ERROR: Cannot perform task, Because: %s\n", err.Error())
		fmt.Println("ERROR: Stopping all process and exiting")
		os.Exit(1)
	}

	fmt.Println("INFO: Performed all tasks successfully")
}
