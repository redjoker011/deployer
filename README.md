# go-deploy - A Go (golang) cli deployer
Distributable build deployer written in go

<!-- MarkdownTOC -->

- [Overview](#overview)
- [Features](#features)
- [Usage](#usage)

<!-- /MarkdownTOC -->

## Overview

Go-Deploy is a cli program designed to deploy distributable build into host
server.

Install it with:

    $ go get redjoker011/go-deploy

or copy build from `pkg/go-deploy` into your project `bin` directory

## Features

- Customizable deployer configuration file.
- Ensure All configuration details were supplied in json file.
- Ensure branch is synced and has no uncommitted changes.
- Ensure all test pass upon deployment.
- Has two available deployment environment `staging` and `production`
- By default use environment configuration file in `deploy` directory
- By default deploy distributable build in `dist` directory

## Usage

Create json configuration file in project root directory inside `deploy`
following environment(staging/production) name.

```
// Deployment Configuration File
// staging.json

{
  "host": "ec2-12-345-678-90.ap-southeast-1.compute.amazonaws.com",
  "user": "ubuntu",
  "branch": "development",
  "deployTo": "/var/www/your_project_name"
}
```

Execute deployer

    $ bin/go-deploy staging

Deploy using different configuration file

    $ bin/go-deploy staging --config="path_to_your_configuration_file"

Deploy using different distributable build

    $ bin/go-deploy staging --dist="path_to_your_distributable_build"
