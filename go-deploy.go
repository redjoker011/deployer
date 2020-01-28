package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var app = cli.NewApp()

func info() {
	app.Name = "Go Deployer"
	app.Usage = "Deploy Project into hosting server"
	app.Version = "1.0.0"
	app.Authors = []*cli.Author{
		&cli.Author{
			Name:  "Peter John Alvarado",
			Email: "redjoker011@gmail.com",
		},
	}
}

type Config struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Branch   string `json:"branch"`
	DeployTo string `json:"deployTo"`
}

func readConfig(filename string) (Config, error) {
	var config Config

	if len(filename) == 0 {
		return config, nil
	}

	file, err := os.Open(filename)

	if err != nil {
		return config, err
	}

	defer file.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(file)

	json.Unmarshal(byteValue, &config)

	return config, nil
}

func main() {
	info()
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "staging",
			Usage: "Deploy on staging server",
			Flags: flags,
			Action: func(c *cli.Context) error {
				file := c.String("config")

				if file == "" {
					file = "deploy/staging.json"
				}

				config, err := readConfig(file)

				branchStatus()
				prevBranch := currentBranch()

				if config.Branch != prevBranch {
					fmt.Printf("Checkout to %s\n", config.Branch)
					checkOut(config.Branch)
				}

				checkUpstream(config.Branch)

				rebaseFromRemoteTree(config.Branch)

				installDependencies()

				invokeUnitTest()

				if err != nil {
					fmt.Println(err)
				}

				return nil
			},
		},
		{
			Name:  "production",
			Usage: "Deploy on production server",
			Flags: flags,
			Action: func(c *cli.Context) error {
				file := c.String("config")

				if file == "" {
					file = "deploy/production.json"
				}

				config, err := readConfig(file)
				fmt.Println("c", config)

				if err != nil {
					fmt.Println(err)
				}

				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func branchStatus() {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`git`, []string{"status", "--porcelain"}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		os.Exit(1)
	}

	if out.String() != "" {
		fmt.Println("Please commit/unstaged your changes to proceed! 💥")
		os.Exit(1)
	}
}

func currentBranch() string {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`git`, []string{"rev-parse", "--abbrev-ref", "HEAD"}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		os.Exit(1)
	}

	return out.String()
}

func checkOut(b string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`git`, []string{"checkout", b}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		os.Exit(1)
	}
}

func checkUpstream(b string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`git`, []string{"log", "origin/" + b + "..HEAD"}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		fmt.Println("\n Please ensure your local branch is synced with your remote branch 💥")
		os.Exit(1)
	}
}

func rebaseFromRemoteTree(b string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`git`, []string{"pull", "--rebase", "origin", b}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		os.Exit(1)
	}

	fmt.Println(out.String())
}

func installDependencies() {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`yarn`, []string{"install"}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		os.Exit(1)
	}

	fmt.Println(out.String())
}

func invokeUnitTest() {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(`yarn`, []string{"test:unit"}...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
		os.Exit(1)
	}

	fmt.Println(out.String())
}
