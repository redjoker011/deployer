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
	"strings"
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

func (c *Config) sshOpts() string {
	return fmt.Sprintf("%s@%s", c.User, c.Host)
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
		&cli.StringFlag{
			Name:  "dist, d",
			Usage: "Distribution `DIRECTORY`",
			Value: "/dist",
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

				build("staging")

				dist := c.String("dist")
				deploy(dist, config)

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
	cmd := "git checkout"
	execute(cmd)
}

func checkUpstream(b string) {
	cmd := fmt.Sprintf("git log origin/%s..HEAD", b)
	execute(cmd)
}

func rebaseFromRemoteTree(b string) {
	cmd := fmt.Sprintf("git pull --rebase origin %s", b)
	execute(cmd)
}

func installDependencies() {
	cmd := fmt.Sprintf("yarn install")
	execute(cmd)
}

func invokeUnitTest() {
	cmd := "yarn test:unit"
	execute(cmd)
}

func build(m string) {
	cmd := fmt.Sprintf("yarn build -m %s", m)
	execute(cmd)
}

func deploy(d string, c Config) {
	dir, err := exec.Command(`date`, []string{"+%s"}...).Output()
	if err != nil {
		log.Fatal(err)
	}
	releasePath := fmt.Sprintf("%s/releases/%s", c.DeployTo, dir)
	// current := fmt.Sprintf("%s/current", c.DeployTo)

	createReleaseDirectory(c, releasePath)
	upload(c, releasePath)
}

func createReleaseDirectory(c Config, path string) {
	err := exec.Command(`ssh`, []string{c.sshOpts(), "mkdir", "-p", "--", c.DeployTo, path, path}...)
	if err != nil {
		log.Fatal(err)
	}
}

func upload(c Config, path string) {
	cmd := fmt.Sprintf("rsync -rv dist/%s:%s", c.sshOpts(), path)
	execute(cmd)
}

func execute(cmdName string) {
	cmdArgs := strings.Fields(cmdName)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdErr := cmd.Run(); cmdErr != nil {
		fmt.Println(fmt.Sprint(cmdErr) + ": " + stderr.String())
	}

	fmt.Println(out.String())
}
