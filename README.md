# GoMinewrap [![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)]() [![Chat on discord](https://img.shields.io/badge/chat%20on-discord-yellow.svg)](https://discord.gg/tae9mst) [![PayPal me!](https://img.shields.io/badge/PayPal-me-lightgrey.svg)](https://www.paypal.me/SuperPykkon)
GoMinewrap is a wrapper for your Minecraft server console. It comes with many new features that could possibly make server management easier. Released under MIT license.  
It allows you to control multiple server, switch between them, and even back them up. It includes a web interface with a modern console log with highlighting, the web interface supports multiple users too.  

## Prerequisites
* [Go](https://golang.org) 1.8

## Cloning and Building
1. `go get github.com/SuperPykkon/GoMinewrap`  
2. `cd $GOPATH/src/github.com/SuperPykkon/GoMinewrap`  
3. `go run *.go`  
Pre-compiled executables are available on the releases tab.

## Running
You will need to configure the configuration under `GoMinewrap/config.yml` to set the servers for the wrapper. You can manually set the config path using a flag. By default, the web server runs on port 25465. For help on wrapper commands, type !help on the command line.
