# GoMinewrap [![Go Report](https://goreportcard.com/badge/github.com/SuperPykkon/GoMinewrap)](https://goreportcard.com/report/github.com/SuperPykkon/GoMinewrap) [![Travis](https://travis-ci.org/SuperPykkon/GoMinewrap.svg?branch=master)]() [![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)]() [![Chat on discord](https://img.shields.io/badge/chat%20on-discord-yellow.svg)](https://discord.gg/tae9mst) [![PayPal me!](https://img.shields.io/badge/PayPal-me-lightgrey.svg)](https://www.paypal.me/SuperPykkon)
GoMinewrap is a wrapper for your Minecraft server console. It comes with many new features that could possibly make server management easier. Released under MIT license.  
It allows you to control multiple server, switch between them, and even back them up. It includes a web interface with a modern console log with highlighting, the web interface supports multiple users too.  

## Prerequisites
* [Go](https://golang.org) 1.8

## Cloning and Building
1. `go get github.com/SuperPykkon/GoMinewrap`  
2. `cd $GOPATH/src/github.com/SuperPykkon/GoMinewrap`  
3. `go run mcs.go`  
Pre-compiled executables are available on the [releases tab](https://github.com/SuperPykkon/GoMinewrap/releases).  
Executables available for Windows, Linux and Mac.

## Screenshots
* http://imgur.com/gallery/0ghEC

## Features
* Completely changes the way the console logs are displayed.
* Multi server support. You can add and run as many servers as you want.
* Real time web based console with authentication which has multi user support.
* A fully functional backup command. You can backup a single or all server (s) with one simple command.
* Very customizable. You can change almost everything to your likings on the config file.
... And much more!

## Running
You will need to configure the configuration under `GoMinewrap/config.yml` to set the servers for the wrapper. You can manually set the config path using a flag. By default, the web server runs on port 80. For help on wrapper commands, type !help on the command line.
