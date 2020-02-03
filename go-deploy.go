package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
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

var file string
var mode string

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
				file = c.String("config")
				mode = "staging"

				if file == "" {
					file = fmt.Sprintf("deploy/%s.json", mode)
				}

				config, err := readConfig(file)
				if err != nil {
					fmt.Println(err)
				}

				branchStatus()
				prevBranch := currentBranch()
				checkBranch(config.Branch, prevBranch)
				checkUpstream(config.Branch)
				rebaseFromRemoteTree(config.Branch)
				installDependencies()
				invokeUnitTest()
				build(mode)
				dist := c.String("dist")
				deploy(dist, config)
				checkBranch(prevBranch, config.Branch)

				return nil
			},
		},
		{
			Name:  "production",
			Usage: "Deploy on production server",
			Flags: flags,
			Action: func(c *cli.Context) error {
				file = c.String("config")
				mode = "production"

				if file == "" {
					file = fmt.Sprintf("deploy/%s.json", mode)
				}

				config, err := readConfig(file)
				if err != nil {
					fmt.Println(err)
				}

				branchStatus()
				prevBranch := currentBranch()
				checkBranch(config.Branch, prevBranch)
				checkUpstream(config.Branch)
				rebaseFromRemoteTree(config.Branch)
				installDependencies()
				invokeUnitTest()
				build(mode)
				dist := c.String("dist")
				deploy(dist, config)
				checkBranch(prevBranch, config.Branch)

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

func checkBranch(crnt string, prev string) {
	if crnt != prev {
		fmt.Printf("Checkout to %s\n", crnt)
		checkOut(crnt)
	}
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
	current := fmt.Sprintf("%s/current", c.DeployTo)

	createReleaseDirectory(c, releasePath)
	upload(c, releasePath)
	updateSymbolicLink(c, releasePath, current)
	done()
}

func createReleaseDirectory(c Config, path string) {
	fmt.Println("\nPreparing release directory....\n")
	err := exec.Command(`ssh`, []string{c.sshOpts(), "mkdir", "-p", "--", c.DeployTo, path, path}...)
	if err != nil {
		log.Fatal(err)
	}
}

func upload(c Config, path string) {
	fmt.Printf("\nUploading build from %s...\n", path)
	cmd := fmt.Sprintf("rsync -rv dist/%s:%s", c.sshOpts(), path)
	execute(cmd)
}

func updateSymbolicLink(c Config, path string, current string) {
	fmt.Printf("\nPointing release to %s...\n", path)
	cmd := fmt.Sprintf("ssh %s rm -rf %s; ln -s %s %s", c.sshOpts(), current, path, current)
	execute(cmd)
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func execute(cmdName string) {
	cmdArgs := strings.Fields(cmdName)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)

	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()

	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("Failed to capture stdout or stderr\n")
	}
	outStr, _ := string(stdout), string(stderr)
	fmt.Printf("\n%s", outStr)
}

func done() {
	fmt.Println("Deployment Done 🎉")
}
