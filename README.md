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

or copy build from `pkg/go-deploy`

## Features

- Customizable deployer configuration file.
- Ensure branch is synced and has no uncommitted changes.
- Ensure all test pass upon deployment.
- Has two available deployment environment `staging` and `production`

## Usage

Go-Deploy command

```
$ go-deploy environment options
```
