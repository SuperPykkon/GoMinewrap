package main

import (
    "fmt"
    "strings"
    "strconv"
    "os"
    "os/exec"
    "regexp"
    "time"
    "io"
    "bufio"
    "net/http"
    "html/template"
    "context"
    "flag"
    "sync"
    "runtime"
    "text/tabwriter"

    "github.com/fatih/color"
    "github.com/gorilla/websocket"
    "github.com/dgrijalva/jwt-go"
    "github.com/spf13/viper"
)

const (
    clrRed = "\x1b[31;22m"
    clrDarkRed = "\x1b[31m"
    clrGreen = "\x1b[32;22m"
    clrDarkGreen = "\x1b[32m"
    clrYellow = "\x1b[33;22m"
    clrDarkYellow = "\x1b[33m"
    clrBlue = "\x1b[34;22m"
    clrDarkBlue = "\x1b[34m"
    clrMagenta = "\x1b[35;22m"
    clrDarkMagenta = "\x1b[35m"
    clrCyan = "\x1b[36;22m"
    clrDarkCyan = "\x1b[36m"
    clrWhite = "\x1b[37;22m"
    clrGray = "\x1b[37m"
    clrDarkGray = "\x1b[38;22m"
    clrEnd = "\x1b[0m"
)

type Key int
const MyKey Key = 0

var (
    enableFilters bool = true
    logWarnSpacer bool
    logWebconWarnSpacer bool
    logErrorSpacer  bool
    logWebconErrorSpacer bool
)

type Claims struct {
    Username string `json:"username"`
    // recommended having
    jwt.StandardClaims
}

type wsExec struct {
    Token string `json: "token"`
    Command string `json: "command"`
}

type consoleTemplate struct {
    Username string
    Token string
    Server string
    Name string
    Servers map[string] interface{}
}

type consoleLogin struct {
    Status string
}

type wsUserdata struct {
    Conn *websocket.Conn
    IP string
    Port string
    Username string
    Server string
}
var wsConns = make(map[int] wsUserdata)

type slogs struct {
    Type string
    Log string
}

type server struct {
    Process *exec.Cmd
    StdinPipe io.WriteCloser
    StdoutPipe io.ReadCloser
    PID int
    Logs map[int] slogs
    Status string
}
var servers = make(map[string] server)
var activeServer string

var config, configDir string
var users map [string] string
var wg sync.WaitGroup

func main() {
    users = make(map[string] string)

    flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  	flags.StringVar(&config, "config", "config", "Name of config file (without extension)")
    flags.StringVar(&configDir, "configDir", "GoMinewrap/", "Path to look for the config file in")
  	flags.Parse(os.Args[1:])

    viper.SetConfigName(config)
    viper.AddConfigPath(configDir)
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
			  fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "VIPER" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing config file or invalid path." + clrEnd)
        os.Exit(1)
		}

    fmt.Println("/*\n *    GoMinewrap v" + viper.GetString("version") + " by SuperPykkon.\n */")

    if viper.GetBool("webcon.enabled") {
        for _, u := range viper.Get("webcon.users").([]interface{}) {
            users[strings.Split(u.(string), ":")[0]] = strings.Split(u.(string), ":")[1]
        }
        go webconHandler()
    }
    enableFilters = viper.GetBool("server.filters.enabled")
    for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
        if s == viper.GetString("server.primary") {
            activeServer = s
        }
    }
    if activeServer == "" {
        fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "VIPER" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: Invalid primary server '" + clrYellow + viper.GetString("server.primary") + clrRed + "'" + clrEnd)
        os.Exit(1)
    }

    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrDarkYellow + "Attempting to start " + clrDarkCyan + strconv.Itoa(len(viper.Get("server.servers").(map[string]interface{}))) + clrDarkYellow + " servers\n                           with the primary server: " + clrDarkCyan + viper.GetString("server.primary") + clrEnd)
    go func() {
        input := bufio.NewReader(os.Stdin)
        for {
            command, _ := input.ReadString('\n')
            serverCommandHandler(command)
        }
    } ()

    serverRun("*")
    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrDarkYellow + "All servers are closed." + clrEnd)
    fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrGreen + "Exiting GoMinewrap, thank you and good bye." + clrEnd)
}


/*

          Minecraft server

*/

func serverRun(server string) {
    if server == "*" {
        for name, _ := range viper.Get("server.servers").(map[string] interface{}) {
            if viper.GetBool("server.servers." + name + ".enabled") {
                wg.Add(1)
                go serverHandler(name)
                time.Sleep(time.Millisecond * 1000)
            }
        }
    } else {
        wg.Add(1)
        go serverHandler(server)
    }
    wg.Wait()
}

func serverHandler(name string) bool {
    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrDarkYellow + "Server starting: " + clrDarkCyan + name + clrEnd)

    process := exec.Command(strings.Fields(viper.GetString("server.servers." + name + ".startup"))[0], strings.Fields(viper.GetString("server.servers." + name + ".startup"))[1:]...)
    process.Dir = viper.GetString("server.base") + viper.GetString("server.servers." + name + ".dir")
    var status string = "Running"

    stdin, err := process.StdinPipe()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error creating StdinPipe for the process:\n" + err.Error())
        status = "Failed"
    }
    defer stdin.Close()

    stdout, err := process.StdoutPipe()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for the process:\n" + err.Error())
        status = "Failed"
    }

    output := bufio.NewScanner(stdout)
    go func() {
        for output.Scan() {
            if output.Text() != "" {
                if name == activeServer {
                    fmt.Fprintln(color.Output, filters(output.Text() + "\n", "main"))
                }
                servers[name].Logs[len(servers[name].Logs)] = slogs{Type: "server", Log: output.Text()}
                for _, ud := range wsConns {
                    if ud.Server == name {
                        ud.Conn.WriteJSON(filters(output.Text(), "webcon"))
                    }
                }
            }
        }
    } ()

    if err := process.Start(); err != nil {
        fmt.Fprintln(os.Stderr, "Error: Failed to start the server: " + name + "\n" + err.Error())
        status = "Failed"
    }

    server_ := server{Process: process, StdinPipe: stdin, StdoutPipe: stdout, PID: process.Process.Pid, Logs: make(map[int] slogs), Status: status}
    servers[name] = server_

    if err := process.Wait(); err != nil {
        fmt.Fprintln(os.Stderr, "Error: Failed to wait for the process on the server: " + name + "\n" + err.Error())
    }
    server_.Status = "Stopped"
    servers[name] = server_

    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrDarkYellow + "Server closed: " + clrDarkCyan + name + clrEnd)

    wg.Done()
    return true
}

func serverCommandHandler(command string) {
    command = strings.TrimSpace(command)
    if strings.HasPrefix(command, "!") {
        if len(command) > 1 {
            cmd:for {
                switch strings.Fields(strings.TrimPrefix(command, "!"))[0] {
                    case "help":
                        fmt.Fprintln(color.Output, clrYellow + "==================== " + clrGreen + "GoMinewrap" + clrYellow + " ====================")
                        fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!help" + clrWhite + ":  " + clrDarkCyan + "Display this help menu." + clrEnd)
                        fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!version" + clrWhite + ":  " + clrDarkCyan + "Display the version of GoMinewrap" + clrEnd)
                        fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!filters" + clrWhite + ":  " + clrDarkCyan + "Enable/disable the custom filters" + clrEnd)
                        fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!clear" + clrWhite + ":  " + clrDarkCyan + "Clear the terminal." + clrEnd)
                        fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!server" + clrWhite + ":  " + clrDarkCyan + "Manage all the servers." + clrEnd)
                        fmt.Fprintln(color.Output, clrDarkCyan + "          Note: '*' will execute the command on every single server." + clrEnd)
                        fmt.Fprintln(color.Output, "          " + clrWhite + "- " + clrDarkYellow + "!server [server]" + clrWhite + ":  " + clrDarkCyan + "Switch to a different server." + clrEnd)
                        fmt.Fprintln(color.Output, "          " + clrWhite + "- " + clrDarkYellow + "!server [server] start" + clrWhite + ":  " + clrDarkCyan + "Start a server that's offline." + clrEnd)
                        fmt.Fprintln(color.Output, "          " + clrWhite + "- " + clrDarkYellow + "!server [server or *] stop" + clrWhite + ":  " + clrDarkCyan + "Stop one or all the server(s)." + clrEnd)
                        // fmt.Fprintln(color.Output, "          " + clrWhite + "- " + clrDarkYellow + "!server [server or *] restart" + clrWhite + ":  " + clrDarkCyan + "Restart one or all the server(s)." + clrEnd)
                        fmt.Fprintln(color.Output, "          " + clrWhite + "- " + clrDarkYellow + "!server [server or *] backup" + clrWhite + ":  " + clrDarkCyan + "Backup all the server files." + clrEnd)
                        fmt.Fprintln(color.Output, "          " + clrWhite + "- " + clrDarkYellow + "!server [server or *] exec [command]" + clrWhite + ":  " + clrEnd)
                        fmt.Fprintln(color.Output, clrDarkCyan + "                Execute a command on any server without having to switching." + clrEnd)
                        break cmd

                    case "version":
                        fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap v" + viper.GetString("version") + " by SuperPykkon." + clrEnd)
                        break cmd

                    case "filters":
                        if len(strings.Fields(command)) == 2 {
                            if strings.Fields(command)[1] == "on" {
                                fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrDarkCyan + "Turned " + clrGreen + "on" + clrDarkCyan + " custom log filtering." + clrEnd)
                                enableFilters = true
                            } else if strings.Fields(command)[1] == "off" {
                                fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrDarkCyan + "Turned " + clrRed + "off" + clrDarkCyan + " custom log filtering." + clrEnd)
                                enableFilters = false
                            } else {
                                fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrDarkCyan + "Unknown arguement '" + strings.Fields(command)[1] + "'. !filters [on/off]" + clrEnd)
                            }
                        } else {
                            fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrDarkCyan + "Usage: !filters [on/off]" + clrEnd)
                        }
                        break cmd

                    case "server":
                        if len(strings.Fields(command)) == 1 {
                            fmt.Fprintln(color.Output, clrDarkYellow + "\nCurrently viewing the server " + clrGreen + activeServer + clrDarkYellow + "'s console." + clrEnd)
                            fmt.Fprintln(color.Output, clrDarkYellow + "To switch to a different server, use " + clrMagenta + "!server [server]" + clrEnd)
                            fmt.Fprintln(color.Output, clrDarkYellow + "Available servers: " + clrEnd + "\n")
                            serversTable := tabwriter.NewWriter(os.Stdout, 0, 0, 6, ' ', tabwriter.AlignRight)
                            fmt.Fprintln(serversTable, "SERVER \t STATUS \t PID \t")
                            fmt.Fprintln(serversTable, "------------ \t ------------ \t ------------ \t")
                            for s, sd := range servers {
                                fmt.Fprintln(serversTable, s + " \t " + sd.Status + " \t " + strconv.Itoa(sd.PID) + " \t")
                            }
                            serversTable.Flush()
                        } else if len(strings.Fields(command)) == 2 {
                            for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
                                if s == string(strings.Fields(command)[1]) {
                                    activeServer = strings.Fields(command)[1]
                                    serverCommandHandler("!clear")
                                    for i := 0; i < len(servers[activeServer].Logs); i++ {
                                        fmt.Fprintln(color.Output, filters(servers[activeServer].Logs[i].Log + "\n", "main"))
                                    }
                                    break cmd
                                }
                            }
                            fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrRed + "Invalid server '" + string(strings.Fields(command)[1]) + "'" + clrEnd)
                        } else if len(strings.Fields(command)) >= 3 {
                            if strings.Fields(command)[2] == "exec" {
                                if len(strings.Fields(command)) > 3 {
                                    if strings.Fields(command)[1] == "*" {
                                        for _, sd := range servers {
                                            io.WriteString(sd.StdinPipe, strings.Join(strings.Fields(command)[3:], " ")  + "\n")
                                        }
                                    } else {
                                        for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
                                            if s == string(strings.Fields(command)[1]) {
                                                io.WriteString(servers[string(strings.Fields(command)[1])].StdinPipe, strings.Join(strings.Fields(command)[3:], " ") + "\n")
                                                break cmd
                                            }
                                        }
                                        fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrRed + "Invalid server '" + string(strings.Fields(command)[1]) + "'" + clrEnd)
                                    }
                                }
                            } else if strings.Fields(command)[2] == "stop" {
                                serverCommandHandler("!server " + string(strings.Fields(command)[1]) + " exec stop")
                            } else if strings.Fields(command)[2] == "kill" {
                                if err := servers[string(strings.Fields(command)[1])].Process.Process.Kill(); err != nil {
                                    fmt.Println("Failed to kill process.")
                                }
                            } else if strings.Fields(command)[2] == "start" {
                                for name, _ := range servers {
                                    if name == strings.Fields(command)[1] {
                                        if servers[string(strings.Fields(command)[1])].Status == "Stopped" {
                                            go serverRun(strings.Fields(command)[1])
                                        } else {
                                            fmt.Println("Server is already running.")
                                        }
                                        break cmd
                                    }
                                }
                            /*
                            // FIXME: Restart is totally broken.

                            } else if strings.Fields(command)[2] == "restart" {
                                if strings.Fields(command)[1] == "*" {

                                } else {
                                    io.WriteString(servers[string(strings.Fields(command)[1])].StdinPipe, "stop\n")
                                    svrrestart:for {
                                        if servers[strings.Fields(command)[1]].Status == "Stopped" {
                                            fmt.Println("Sending start command ~")
                                            go serverRun(strings.Fields(command)[1])
                                            break svrrestart
                                        }
                                        //time.Sleep(time.Millisecond * 1000)
                                    }
                                }
                            */
                            } else if strings.Fields(command)[2] == "backup" {
                                t := time.Now().Format("2006-01-02_15.04.05")
                                var d *exec.Cmd
                                if runtime.GOOS == "windows" {
                                    d = exec.Command("cmd", "/c", "mkdir", t)
                                } else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
                                    d = exec.Command("mkdir", t)
                                }
                                d.Dir = viper.GetString("server.backup.dir")
                                d.Run()

                                if strings.Fields(command)[1] == "*" {
                                    fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrGreen + "Begin backup of the server: " + clrYellow + "all servers" + clrGreen +  "." + clrEnd)
                                    for name, _ := range viper.Get("server.servers").(map[string]interface{}) {
                                        fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrGreen + "  ... backing up the server: " + clrYellow + name + clrGreen +  "." + clrEnd)
                                        if runtime.GOOS == "windows" {
                                            cmd := exec.Command("cmd", "/c", "robocopy", "../../" + viper.GetString("server.base") + viper.GetString("server.servers." + name + ".dir"), viper.GetString("server.servers." + name + ".dir"), "/MIR")
                                              cmd.Dir = viper.GetString("server.backup.dir") + t
                                            cmd.Run()
                                        } else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
                                            cmd := exec.Command("cp", "-a", "../../" + viper.GetString("server.base") + viper.GetString("server.servers." + name + ".dir") + ".", viper.GetString("server.servers." + name + ".dir"))
                                            cmd.Dir = viper.GetString("server.backup.dir") + t
                                            cmd.Run()
                                        }
                                    }
                                } else {
                                    fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrGreen + "Begin backup of the server: " + clrYellow + string(strings.Fields(command)[1]) + clrGreen +  "." + clrEnd)
                                    if runtime.GOOS == "windows" {
                                        cmd := exec.Command("cmd", "/c", "robocopy", "../../" + viper.GetString("server.base") + viper.GetString("server.servers." + string(strings.Fields(command)[1]) + ".dir"), viper.GetString("server.servers." + string(strings.Fields(command)[1]) + ".dir"), "/MIR")
                                        cmd.Dir = viper.GetString("server.backup.dir") + t
                                        cmd.Run()
                                    } else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
                                        cmd := exec.Command("cp", "-a", "../../" + viper.GetString("server.base") + viper.GetString("server.servers." + string(strings.Fields(command)[1]) + ".dir") + ".", viper.GetString("server.servers." + string(strings.Fields(command)[1]) + ".dir"))
                                        cmd.Dir = viper.GetString("server.backup.dir") + t
                                        cmd.Run()
                                    }
                                }
                                fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrGreen + "Backup complete: " + clrYellow + viper.GetString("server.backup.dir") + t + " ..." + clrEnd)
                            }
                        }
                        break cmd

                    case "clear":
                        if runtime.GOOS == "windows" {
                            cmd := exec.Command("cmd", "/c", "cls")
                            cmd.Stdout = os.Stdout
                            cmd.Run()
                        } else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
                            cmd := exec.Command("clear")
                            cmd.Stdout = os.Stdout
                            cmd.Run()
                        }
                        break cmd

                    default:
                        fmt.Fprintln(color.Output, clrDarkGray + "[" + clrRed + "*" + clrDarkGray + "]" + clrYellow + " GoMinewrap » " + clrRed + "Unknown command, try '!help' for a list of commands." + clrEnd)
                        break cmd
                }
            }
        }
    } else {
        switch strings.TrimSpace(command) {
            case "stop":
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrMagenta + "Received stop command, exiting the server..." + clrEnd)
                io.WriteString(servers[activeServer].StdinPipe, command + "\n")
                servers[activeServer].Process.Wait()

            case "restart":
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrMagenta + "Received restart command, restarting the server..." + clrEnd)
                io.WriteString(servers[activeServer].StdinPipe, "stop\n")
                servers[activeServer].Process.Wait()

            default:
                io.WriteString(servers[activeServer].StdinPipe, command + "\n")
                servers[activeServer].Process.Wait()
        }
    }
}

func filters(text string, type_ string) string {
    line := strings.TrimSpace(text)
    var (
        logTime bool
        logClrEnd bool
    )
    if type_ == "main" {
        if enableFilters {
            if regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).MatchString(line) {
                line = regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).ReplaceAllString(line, time.Now().Format("15:04:05"))
                logTime = true
            } else { logTime = false }

            if regexp.MustCompile(`INFO\]:`).MatchString(line) {
                if regexp.MustCompile(`\[/[0-9]+(?:\.[0-9]+){3}:[0-9]+\] logged in with entity id \d`).MatchString(line) {
                    line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkGreen + "+" + clrDarkCyan + " INFO:" + clrGreen)
                    logClrEnd = true
                } else if regexp.MustCompile(`[a-zA-Z0-9_.-] lost connection\:`).MatchString(line) {
                    line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkRed + "-" + clrDarkCyan + " INFO:" + clrRed)
                    logClrEnd = true
                } else if regexp.MustCompile(`Done \(\d*\.?\d*s\)! For help, type "help" or "\?"`).MatchString(line) {
                    line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkGreen + "=" + clrGreen + " DONE:" + clrEnd)
                } else {
                    if regexp.MustCompile(`[a-zA-Z0-9_.-] issued server command\:`).MatchString(line) {
                        line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrCyan + "|" + clrDarkCyan + " INFO:" + clrEnd)
                        line = regexp.MustCompile(`\: `).ReplaceAllString(line, ": " + clrCyan)
                        logClrEnd = true
                    } else {
                        line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkCyan + "| INFO:" + clrEnd)
                    }
                }

                if logWarnSpacer {
                    logWarnSpacer = false
                } else if logErrorSpacer {
                    logErrorSpacer = false
                }
            } else if regexp.MustCompile(`WARN\]:`).MatchString(line) {
                line = regexp.MustCompile(`WARN\]:`).ReplaceAllString(line, clrDarkYellow + "! WARN:" + clrYellow)
                logWarnSpacer = true
                logClrEnd = true
            } else if regexp.MustCompile(`ERROR\]:`).MatchString(line) {
                line = regexp.MustCompile(`ERROR\]:`).ReplaceAllString(line, clrDarkRed + "x ERROR:" + clrRed)
                logErrorSpacer = true
                logClrEnd = true
            }

            line = regexp.MustCompile(`\[Server\]`).ReplaceAllString(line, clrMagenta + "[Server]" + clrEnd)

            if logWarnSpacer && !logTime {
                return "              " + clrYellow + line + clrEnd
            } else if logErrorSpacer && !logTime {
                return "              " + clrRed + line + clrEnd
            } else if logClrEnd {
                return line + clrEnd
            } else {
                return line
            }
        } else {
            return line
        }
    } else if type_ == "webcon" {
        if regexp.MustCompile(`INFO\]:`).MatchString(line) {
            line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, "")
            if regexp.MustCompile(`\[/[0-9]+(?:\.[0-9]+){3}:[0-9]+\] logged in with entity id \d`).MatchString(line) {
                line = "<div class='log joined'> <i class='add user icon'></i> " + line + "</div>"
            } else if regexp.MustCompile(`[a-zA-Z0-9_.-] lost connection\:`).MatchString(line) {
                line = "<div class='log left'> <i class='remove user icon'></i> " + line + "</div>"
            } else if regexp.MustCompile(`[a-zA-Z0-9_.-] issued server command\:`).MatchString(line) {
                line = "<div class='log cmd'>" + line + "</div>"
                line = regexp.MustCompile(`\: `).ReplaceAllString(line, ": <span style='color: #55FFFF; margin-left: 4px'>")
                line = line + "</span>"
            } else {
                line = "<div class='log info'>" + line + "</div>"
            }

            if logWebconWarnSpacer {
                logWebconWarnSpacer = false
            } else if logWebconErrorSpacer {
                logWebconErrorSpacer = false
            }
        } else if regexp.MustCompile(`WARN\]:`).MatchString(line) {
            line = regexp.MustCompile(`WARN\]:`).ReplaceAllString(line, "")
            line = "<div class='log warning'> <i class='warning circle icon'></i> " + line + "</div>"
            logWebconWarnSpacer = true
        } else if regexp.MustCompile(`ERROR\]:`).MatchString(line) {
            line = regexp.MustCompile(`ERROR\]:`).ReplaceAllString(line, "")
            line = "<div class='log error'> <i class='remove circle icon'></i> " + line + "</div>"
            logWebconErrorSpacer = true
        } else {
            if logWebconWarnSpacer {
                line = "<div class='log warning log_spacer'>" + line + "</div>"
            } else if logWebconErrorSpacer {
                line = "<div class='log error log_spacer'>" + line + "</div>"
            } else {
                line = "<div class='log undefined'>" + line + "</div>"
            }
        }

        if regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).MatchString(line) {
            line = regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).ReplaceAllString(line, "")
            line = "<div class='row'><span class='time'>" + time.Now().Format("15:04:05") + "</span>" + line
        } else {
            line = "<div class='row'><span class='time'></span>" + line
        }

        line = regexp.MustCompile(`\[Server\]`).ReplaceAllString(line, "<div style='color: #FF55FF; padding: 0px 5px 0px 0px;'>[Server]</div>")
        return line + "</div>"
    }
    return ""
}


/*

          Webcon - Web console

*/


func webconHandler() {
    time.Sleep(time.Millisecond * 1500)

    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrDarkYellow + "Starting the web server on: " + clrDarkCyan + viper.GetString("webcon.host") + ":" + viper.GetString("webcon.port") + clrDarkYellow + " ..." + clrEnd)
    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrDarkYellow + "Loaded " + clrDarkCyan + strconv.Itoa(len(viper.Get("webcon.users").([]interface{}))) + clrDarkYellow + " users for webcon login." + clrEnd)

    http.Handle("/", webconAuthValidate(webconRootHandler))
    http.HandleFunc("/ws", wsHandler)
    http.HandleFunc("/login", webconAuthLogin)
    http.HandleFunc("/logout", webconAuthValidate(webconAuthLogout))
    http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("static/web/html/resources"))))

    http.ListenAndServe(viper.GetString("webcon.host") + ":" + viper.GetString("webcon.port"), nil)
}

func webconRootHandler(w http.ResponseWriter, r *http.Request)  {
    claims, ok := r.Context().Value(MyKey).(Claims)
    if !ok {
        http.Redirect(w, r, "/login", 307)
    }

    if _, err := os.Stat(viper.GetString("webcon.dir") + "index.html"); os.IsNotExist(err) {
        fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing file 'index.html' or invalid path." + clrEnd)
        fmt.Fprintln(w, "An internal error has occured.")
    } else {
        var blocked bool
        var server bool = false
        for _, a := range viper.Get("webcon.blacklist.IP").([]interface{}) {
            if strings.Split(r.RemoteAddr, ":")[0] == a.(string) {
                fmt.Fprintln(w, "You do not have permission to access this page.")
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Terminated connection from blacklisted IP: " + clrYellow + a.(string) + clrEnd)
                blocked = true
            }
        }

        for _, a := range viper.Get("webcon.blacklist.users").([]interface{}) {
            if claims.Username == a.(string) {
                fmt.Fprintln(w, "You do not have permission to access this page.")
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Terminated connection from blacklisted user: " + clrYellow + a.(string) + clrEnd)
                blocked = true
            }
        }

        for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
            if s == strings.Join(r.URL.Query()["server"], " ") {
               server = true
            }
        }

        if !blocked {
            if server || len(r.URL.Query()["server"]) == 0 {
                cookie, _ := r.Cookie("Auth")
                var temp consoleTemplate
                if len(r.URL.Query()["server"]) > 0 {
                    temp = consoleTemplate{Username: claims.Username, Token: cookie.Value, Server: strings.Join(r.URL.Query()["server"], " "), Name: viper.GetString("server.name"), Servers: viper.Get("server.servers").(map[string]interface{})}
                } else {
                    temp = consoleTemplate{Username: claims.Username, Token: cookie.Value, Server: viper.GetString("server.primary"), Name: viper.GetString("server.name"), Servers: viper.Get("server.servers").(map[string]interface{})}
                }

                t := template.Must(template.ParseFiles(viper.GetString("webcon.dir") + "index.html"))
                t.Execute(w, temp)
            } else {
                fmt.Fprintln(w, "Invalid server entered.")
            }
        }
    }
}

func webconAuthLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        if _, err := os.Stat(viper.GetString("webcon.dir") + "login.html"); os.IsNotExist(err) {
            fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing file 'login.html' or invalid path." + clrEnd)
            fmt.Fprintln(w, "An internal error has occured.")
        } else {
            var blocked bool = false
            for _, a := range viper.Get("webcon.blacklist.IP").([]interface{}) {
                if strings.Split(r.RemoteAddr, ":")[0] == a.(string) {
                    fmt.Fprintln(w, "You do not have permission to access this page.")
                    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Terminated connection from blacklisted IP: " + clrYellow + a.(string) + clrEnd)
                    blocked = true
                }
            }

            if !blocked {
                t := template.Must(template.ParseFiles(viper.GetString("webcon.dir") + "login.html"))
                temp := consoleLogin{Status: ""}
                t.Execute(w, temp)
            }
        }
    } else {
        r.ParseForm()
        if users[r.Form["username"][0]] != "" && users[r.Form["username"][0]] == r.Form["password"][0] {
            if viper.GetBool("webcon.messages.login_success") { fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrGreen + "Authentication successful from: " + clrYellow + strings.Split(r.RemoteAddr, ":")[0] + clrGreen + "\n                           user: " + clrYellow + r.Form["username"][0] + clrEnd) }
            // Expires the token and cookie in 1 hour
            expireToken := time.Now().Add(time.Hour * 1).Unix()
            expireCookie := time.Now().Add(time.Hour * 1)

            // We'll manually assign the claims but in production you'd insert values from a database
            claims := Claims {
                r.Form["username"][0],
                jwt.StandardClaims {
                    ExpiresAt: expireToken,
                    Issuer:    "GoMinewrap",
                },
            }

            // Create the token using your claims
            token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

            // Signs the token with a secret.
            signedToken, _ := token.SignedString([]byte("secret"))

            // Place the token in the client's cookie
            cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true}
            http.SetCookie(w, &cookie)

            // Redirect the user to his profile
            http.Redirect(w, r, "/", 307)
        } else {
            if _, err := os.Stat(viper.GetString("webcon.dir") + "login.html"); os.IsNotExist(err) {
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing file 'login.html' or invalid path." + clrEnd)
                fmt.Fprintln(w, "An internal error has occured.")
            } else {
                if viper.GetBool("webcon.messages.login_fail") { fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Authentication failed from: " + clrYellow + strings.Split(r.RemoteAddr, ":")[0] + clrRed + "\n                           reason: Invalid username/password." + clrEnd) }
                t := template.Must(template.ParseFiles(viper.GetString("webcon.dir") + "login.html"))
                temp := consoleLogin{Status: "Invalid username or password."}
                t.Execute(w, temp)
            }
        }
    }
}

// middleware to protect private pages
func webconAuthValidate(page http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){

        // If no Auth cookie is set then return a 404 not found
        cookie, err := req.Cookie("Auth")
        if err != nil {
            http.Redirect(res, req, "/login", 307)
            return
        }

        // Return a Token using the cookie
        token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error){
            // Make sure token's signature wasn't changed
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("Unexpected siging method")
            }
            return []byte("secret"), nil
        })
        if err != nil {
            http.Redirect(res, req, "/login", 307)
            return
        }

        // Grab the tokens claims and pass it into the original request
        if claims, ok := token.Claims.(*Claims); ok && token.Valid {
            ctx := context.WithValue(req.Context(), MyKey, *claims)
            page(res, req.WithContext(ctx))
        } else {
            http.Redirect(res, req, "/login", 307)
            return
        }
    })
}

func webconAuthLogout(w http.ResponseWriter, r *http.Request) {
    deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
    http.SetCookie(w, &deleteCookie)
    http.Redirect(w, r, "/login", 307)
}


/*

          Webcon Websocket

*/


func wsHandler(w http.ResponseWriter, r *http.Request) {
  	if r.Header.Get("Origin") != "http://" + r.Host {
    		http.Error(w, "Origin not allowed", 403)
    		return
  	}
  	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
  	if err != nil { http.Error(w, "Could not open websocket connection", http.StatusBadRequest) }

    token, err := jwt.ParseWithClaims(strings.Join(r.URL.Query()["token"], " "), &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Make sure token's signature wasn't changed
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected siging method")
        }
        return []byte("secret"), nil
    })
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        CID := len(wsConns)
        wsConns[CID] = wsUserdata{Conn: conn, IP: strings.Split(r.RemoteAddr, ":")[0], Port: strings.Split(r.RemoteAddr, ":")[1], Username: claims.Username, Server: strings.Join(r.URL.Query()["server"], " ")}
      	go wsConnectionHandler(conn, strings.Split(r.RemoteAddr, ":")[0], CID, strings.Join(r.URL.Query()["server"], " "))
    }
}

func wsConnectionHandler(conn *websocket.Conn, IP string, CID int, server string) {
    if viper.GetBool("webcon.messages.ws_connect") { fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ") " + clrMagenta + "WS" + clrDarkCyan + ": " + clrGreen + "Connection established from: " + clrYellow + IP + clrDarkCyan + " [" + strconv.Itoa(len(wsConns)) + "]" + clrEnd) }
    for {
        exec := wsExec{}
        err := conn.ReadJSON(&exec)
        if err != nil {
            delete(wsConns, CID)
            if viper.GetBool("webcon.messages.ws_disconnect") { fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ") " + clrMagenta + "WS" + clrDarkCyan + ": " + clrRed + "Connection terminated from: " + clrYellow + IP + clrDarkCyan + " [" + strconv.Itoa(len(wsConns)) + "]" + clrEnd) }
            break
        }
        token, err := jwt.ParseWithClaims(exec.Token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            // Make sure token's signature wasn't changed
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("Unexpected siging method")
            }
            return []byte("secret"), nil
        })
        if claims, ok := token.Claims.(*Claims); ok && token.Valid {
            if exec.Command == "/ws-gh" {
                for i := 0; i <= len(servers[server].Logs); i++ {
                    if servers[server].Logs[i].Type == "server" {
                        conn.WriteJSON(filters(servers[server].Logs[i].Log, "webcon"))
                    }
                }
            } else {
                if server == activeServer {
                    if viper.GetBool("webcon.messages.exec_command") { fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrYellow + claims.Username + clrEnd + " executed a server command: " + clrCyan + "/" + exec.Command + clrEnd) }
                }
                if viper.GetBool("webcon.messages.exec_command") { servers[server].Logs[len(servers[server].Logs)] = slogs{Type: "webcon", Log: time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrYellow + claims.Username + clrEnd + " executed a server command: " + clrCyan + "/" + exec.Command + clrEnd} }

                time.Sleep(time.Millisecond)
                io.WriteString(servers[server].StdinPipe, exec.Command + "\n")
            }
        }
    }
}
