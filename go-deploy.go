package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
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
				fmt.Println("c", config)

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
