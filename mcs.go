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
    "os/signal"
    "flag"

    "github.com/fatih/color"
    "github.com/gorilla/websocket"
    "github.com/dgrijalva/jwt-go"
)

const version = "0.1"
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
}

type consoleLogin struct {
    Status string
}

type wsUserdata struct {
    Conn *websocket.Conn
}

type processData struct {
    Process *exec.Cmd
    StdinPipe io.WriteCloser
    StdoutPipe io.ReadCloser
}

var runCmd, serverDir, webcon, webconDir, webconHost, webconPort, webconUser, webconPass string
var logs map[int] string
var users map [string] string
var wsConns = make(map[string] wsUserdata)
var servers = make(map[string] processData)

func main() {
    logs = make(map[int] string)
    users = make(map[string] string)

    flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  	flags.StringVar(&runCmd, "runCmd", "java -Xmx1G -jar spigot.jar", "Startup script for the Minecraft server.")
    flags.StringVar(&serverDir, "serverDir", "server", "Directory of the server.")
  	flags.StringVar(&webcon, "webcon", "off", "Enable or disable webcon (web console).")
    flags.StringVar(&webconDir, "webconDir", "static/web/html/", "Directory of the webcon files.")
    flags.StringVar(&webconHost, "webconHost", "", "Enable or disable webcon (web console).")
    flags.StringVar(&webconPort, "webconPort", "80", "Which port will webcon listen on?")
    flags.StringVar(&webconUser, "webconUser", "admin", "Username for webcon login.")
    flags.StringVar(&webconPass, "webconPass", "changeme", "Password for webcon login.")
  	flags.Parse(os.Args[1:])

    users[webconUser] = webconPass

    fmt.Println("/*\n *    GoMinewrap v" + version + " by SuperPykkon.\n */")

    if webcon == "on" { go webconHandler() }

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
        for sig := range c {
            fmt.Println(sig)
            io.WriteString(servers["main"].StdinPipe, "stop\n")
        }
    }()

    run()
}


/*

          Minecraft server

*/


func run() {
    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrEnd + "Starting the server...")
    process := exec.Command(strings.Fields(runCmd)[0], strings.Fields(runCmd)[1:]...)
    process.Dir = serverDir
    var restart bool

    stdin, err := process.StdinPipe()
    if err != nil { fmt.Fprintln(os.Stderr, "Error creating StdinPipe for the process:\n" + err.Error()) }
    defer stdin.Close()

    stdout, err := process.StdoutPipe()
    if err != nil { fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for the process:\n" + err.Error()) }

    servers["main"] = processData{Process: process, StdinPipe: stdin, StdoutPipe: stdout}
    output := bufio.NewScanner(stdout)

    go func() {
        for output.Scan() {
            if output.Text() != "" {
                fmt.Fprintln(color.Output, filters(output.Text() + "\n", "main"))
                for _, conn := range wsConns {
                    conn.Conn.WriteJSON(filters(output.Text(), "webcon"))
                }
                logs[len(logs)] = output.Text()
            }
        }
    } ()

    input := bufio.NewReader(os.Stdin)
    go func() {
        for {
            command, _ := input.ReadString('\n')
            if strings.HasPrefix(command, "!") {
                if len(strings.TrimSpace(command)) > 1 {
                    switch strings.Fields(strings.TrimSpace(strings.TrimPrefix(command, "!")))[0] {
                        case "help":
                            fmt.Fprintln(color.Output, clrYellow + "==================== " + clrGreen + "GoMinewrap" + clrGreen + " ====================")
                            fmt.Fprintln(color.Output, "     " + clrDarkMagenta + "GoMinewrap" + clrWhite + ":" + clrEnd)
                            fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!help" + clrWhite + ":  " + clrDarkCyan + "Display this help menu." + clrEnd)
                            fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!version" + clrWhite + ":  " + clrDarkCyan + "Display the version of GoMinewrap" + clrEnd)
                            fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "!filters" + clrWhite + ":  " + clrDarkCyan + "Enable/disable the custom filters" + clrEnd)
                            fmt.Println("\n")
                            fmt.Fprintln(color.Output, "     " + clrDarkMagenta + "Server" + clrWhite + ":" + clrEnd)
                            fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "restart" + clrWhite + ":  " + clrDarkCyan + "Properly restart the server." + clrEnd)
                            fmt.Fprintln(color.Output, "     " + clrWhite + "- " + clrDarkYellow + "stop" + clrWhite + ":  " + clrDarkCyan + "Properly exit the server and wrapper." + clrEnd)

                        case "version":
                            fmt.Fprintln(color.Output, clrDarkCyan + "// " + clrYellow + "GoMinewrap v" + version + " by SuperPykkon." + clrEnd)

                        case "filters":
                            if len(strings.Fields(command)) == 2 {
                                if strings.Fields(command)[1] == "on" {
                                    fmt.Fprintln(color.Output, clrDarkCyan + "// " + clrYellow + "GoMinewrap > " + clrDarkCyan + "Turned on custom log filtering." + clrEnd)
                                    enableFilters = true
                                } else if strings.Fields(command)[1] == "off" {
                                    fmt.Fprintln(color.Output, clrDarkCyan + "// " + clrYellow + "GoMinewrap > " + clrDarkCyan + "Turned off custom log filtering." + clrEnd)
                                    enableFilters = false
                                } else {
                                    fmt.Fprintln(color.Output, clrDarkCyan + "// " + clrYellow + "GoMinewrap > " + clrDarkCyan + "Unknown arguement '" + strings.Fields(command)[1] + "'. !filters [on/off]" + clrEnd)
                                }
                            } else {
                                fmt.Fprintln(color.Output, clrDarkCyan + "// " + clrYellow + "GoMinewrap > " + clrDarkCyan + "Usage: !filters [on/off]" + clrEnd)
                            }
                        default:
                            fmt.Fprintln(color.Output, clrDarkCyan + "// " + clrYellow + "GoMinewrap > " + clrRed + "Unknown command, try '!help' for a list of commands." + clrEnd)
                    }
                }
            } else {
                switch strings.TrimSpace(command) {
                    case "stop":
                        fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrEnd + "Received stop command, exiting the server...")
                        io.WriteString(stdin, command + "\n")
                        process.Wait()

                    case "restart":
                        fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "SERVER" + clrWhite + ")" + clrDarkCyan + ": " + clrEnd + "Received restart command, restarting the server...")
                        io.WriteString(stdin, command + "\n")
                        process.Wait()
                        restart = true

                    default:
                        io.WriteString(stdin, command + "\n")
                        process.Wait()
                }
            }
        }
    } ()

    if err = process.Start(); err != nil { fmt.Println(err) }
    if err = process.Wait(); err != nil { fmt.Println(err) }

    if restart { run() }
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

    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrEnd + "Starting the web server...")
    http.Handle("/", webconAuthValidate(webconRootHandler))
    http.HandleFunc("/ws", wsHandler)
    http.HandleFunc("/login", webconAuthLogin)
    http.HandleFunc("/logout", webconAuthValidate(webconAuthLogout))
    http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("static/web/html/resources"))))

    http.ListenAndServe(webconHost + ":" + webconPort, nil)
}

func webconRootHandler(w http.ResponseWriter, r *http.Request)  {
    claims, ok := r.Context().Value(MyKey).(Claims)
    if !ok {
        http.Redirect(w, r, "/login", 307)
    }
    if _, err := os.Stat(webconDir + "index.html"); os.IsNotExist(err) {
        fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing file 'index.html' or invalid path." + clrEnd)
        fmt.Fprintln(w, "An internal error has occured.")
    } else {
        t := template.Must(template.ParseFiles(webconDir + "index.html"))
        cookie, _ := r.Cookie("Auth")
        temp := consoleTemplate{Username: claims.Username, Token: cookie.Value}
        t.Execute(w, temp)
    }
}

func webconAuthLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        if _, err := os.Stat(webconDir + "login.html"); os.IsNotExist(err) {
            fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing file 'login.html' or invalid path." + clrEnd)
            fmt.Fprintln(w, "An internal error has occured.")
        } else {
            t := template.Must(template.ParseFiles(webconDir + "login.html"))
            temp := consoleLogin{Status: ""}
            t.Execute(w, temp)
        }
    } else {
        r.ParseForm()
        if users[r.Form["username"][0]] != "" && users[r.Form["username"][0]] == r.Form["password"][0] {
            fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrGreen + "Authentication successful from: " + strings.Split(r.RemoteAddr, ":")[0] + "\n                           user: " + clrDarkCyan + r.Form["username"][0] + clrEnd)
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
            if _, err := os.Stat(webconDir + "login.html"); os.IsNotExist(err) {
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Error: missing file 'login.html' or invalid path." + clrEnd)
                fmt.Fprintln(w, "An internal error has occured.")
            } else {
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrRed + "Authentication failed from: " + strings.Split(r.RemoteAddr, ":")[0] + "\n                           reason: Invalid username/password." + clrEnd)
                t := template.Must(template.ParseFiles(webconDir + "login.html"))
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
  	if err != nil {
    		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
  	}
    wsConns[strings.Split(r.RemoteAddr, ":")[0]] = wsUserdata{Conn: conn}
  	go wsConnectionHandler(conn, strings.Split(r.RemoteAddr, ":")[0])
}

func wsConnectionHandler(conn *websocket.Conn, IP string) {
    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ") " + clrMagenta + "WS" + clrDarkCyan + ": " + clrGreen + "Connection established from: " + IP + clrDarkCyan + " [" + strconv.Itoa(len(wsConns)) + "]" + clrEnd)
    for {
        exec := wsExec{}
        err := conn.ReadJSON(&exec)
        if err != nil {
            delete(wsConns, IP)
            fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ") " + clrMagenta + "WS" + clrDarkCyan + ": " + clrRed + "Connection terminated from: " + IP + clrDarkCyan + " [" + strconv.Itoa(len(wsConns)) + "]" + clrEnd)
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
                for i := 0; i <= len(logs); i++ {
                    conn.WriteJSON(filters(logs[i], "webcon"))
                }
            } else {
                fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + claims.Username + clrEnd + " executed a server command: " + clrCyan + "/" + exec.Command + clrEnd)
                time.Sleep(time.Millisecond)
                io.WriteString(servers["main"].StdinPipe, exec.Command + "\n")
            }
        }
    }
}
