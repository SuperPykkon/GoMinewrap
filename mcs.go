package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

const (
	clrRed         = "\x1b[31;22m"
	clrDarkRed     = "\x1b[31m"
	clrGreen       = "\x1b[32;22m"
	clrDarkGreen   = "\x1b[32m"
	clrYellow      = "\x1b[33;22m"
	clrDarkYellow  = "\x1b[33m"
	clrBlue        = "\x1b[34;22m"
	clrDarkBlue    = "\x1b[34m"
	clrMagenta     = "\x1b[35;22m"
	clrDarkMagenta = "\x1b[35m"
	clrCyan        = "\x1b[36;22m"
	clrDarkCyan    = "\x1b[36m"
	clrWhite       = "\x1b[37;22m"
	clrGray        = "\x1b[37m"
	clrDarkGray    = "\x1b[38;22m"
	clrEnd         = "\x1b[0m"
)

type Key int

const MyKey Key = 0

var (
	enableFilters        bool = true
	logWarnSpacer        bool
	logWebconWarnSpacer  bool
	logErrorSpacer       bool
	logWebconErrorSpacer bool
)

type Claims struct {
	Username string `json:"username"`
	// recommended having
	jwt.StandardClaims
}

type wsExec struct {
	Token   string `json: "token"`
	Command string `json: "command"`
}

type consoleTemplate struct {
	Username string
	Token    string
	Server   string
}

type consoleLogin struct {
	Status string
}

type wsUserdata struct {
	Conn     *websocket.Conn
	IP       string
	Port     string
	Username string
	Server   string
}

var wsConns = make(map[int]wsUserdata)

type slogs struct {
	Type string
	Log  string
}

type Server struct {
	process    *exec.Cmd
	stdinPipe  io.WriteCloser
	stdoutPipe io.ReadCloser
	logs       map[int]slogs
	status     ServerStatus
}

type ServerStatus int64

const (
	// RUNNING is when the server is up and running
	RUNNING = iota

	// RESTARTING is when the server is up and running
	RESTARTING = iota

	// IDLE is when the server is not running
	IDLE = iota

	// FAILED is when the server failed to start
	FAILED = iota
)

var servers map[string]Server

// Current server screen
var activeServer string

var config, configDir string
var users map[string]string

var wg sync.WaitGroup

func main() {
	users = make(map[string]string)
	servers = make(map[string]Server)

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&config, "config", "config", "Name of config file (without extension)")
	flags.StringVar(&configDir, "configDir", "GoMinewrap/", "Path to look for the config file in")
	flags.Parse(os.Args[1:])

	// Load the config
	viper.SetConfigName(config)
	viper.AddConfigPath(configDir)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"VIPER"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing config file or invalid path."+clrEnd)
		os.Exit(1)
	}

	// Print the program description
	fmt.Println("Running GoMinewrap v" + viper.GetString("version") + " by SuperPykkon released under MIT license.")

	// Run the web console if it is enabled
	if viper.GetBool("webcon.enabled") {
		fmt.Fprintln(color.Output, clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrMagenta+"Starting the web server on: "+viper.GetString("webcon.host")+":"+viper.GetString("webcon.port")+" with "+clrDarkCyan+strconv.Itoa(len(viper.Get("webcon.users").([]interface{})))+clrMagenta+" users loaded."+clrEnd)
		for _, u := range viper.Get("webcon.users").([]interface{}) {
			users[strings.Split(u.(string), ":")[0]] = strings.Split(u.(string), ":")[1]
		}
		go webconHandler()
	}
	// Enable filters is it is set
	enableFilters = viper.GetBool("server.filters.enabled")

	// Load servers
	fmt.Fprintln(color.Output, clrYellow+"Loaded servers: "+clrEnd)

	for name, _ := range viper.Get("server.servers").(map[string]interface{}) {
		if viper.GetBool("server.servers." + name + ".enabled") {
			nprocess := exec.Command(strings.Fields(viper.GetString("server.servers." + name + ".startup"))[0], strings.Fields(viper.GetString("server.servers." + name + ".startup"))[1:]...)
			nprocess.Dir = viper.GetString("server.base") + viper.GetString("server.servers."+name+".dir")
			stdin, _ := nprocess.StdinPipe()
			stdout, _ := nprocess.StdoutPipe()
			servers[name] = Server{process: nprocess, stdinPipe: stdin, stdoutPipe: stdout, logs: make(map[int]slogs), status: IDLE}
			fmt.Fprintln(color.Output, clrWhite+" -  "+clrDarkCyan+name+clrEnd)

		}
	}

	go func() {
		input := bufio.NewReader(os.Stdin)
		for {
			command, _ := input.ReadString('\n')
			serverCommandHandler(command)
		}
	}()

	fmt.Fprintln(color.Output, clrYellow+"Type !help to for help."+clrEnd)

	for true {
	}
}
