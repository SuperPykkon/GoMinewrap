package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

func serverRun() {
	fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"SERVER"+clrWhite+")"+clrDarkCyan+": "+clrMagenta+"Attempting to start "+clrYellow+strconv.Itoa(len(viper.Get("server.servers").(map[string]interface{})))+clrMagenta+" servers..."+clrEnd)
	fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"SERVER"+clrWhite+")"+clrDarkCyan+": "+clrMagenta+"The primary server is: "+clrYellow+viper.GetString("server.primary")+clrEnd)

	wg.Wait()
	fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"SERVER"+clrWhite+")"+clrDarkCyan+": "+clrMagenta+"All servers have been exited."+clrEnd)
}

func serverLogHandler(name string, process *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, status string) {
	output := bufio.NewScanner(stdout)
	go func() {
		for output.Scan() {
			if output.Text() != "" {
				if name == activeServer {
					fmt.Fprintln(color.Output, filters(output.Text()+"\n", "main"))
				}
				servers[name].logs[len(servers[name].logs)] = slogs{Type: "server", Log: output.Text()}
				for _, ud := range wsConns {
					if ud.Server == name {
						ud.Conn.WriteJSON(filters(output.Text(), "webcon"))
					}
				}
			}
		}
	}()

	wg.Done()
}

func serverCommandHandler(command string) {
	command = strings.TrimSpace(command)
	if strings.HasPrefix(command, "!") {
		if len(command) > 1 {
		cmd:
			for {
				switch strings.Fields(strings.TrimPrefix(command, "!"))[0] {
				case "help":
					fmt.Fprintln(color.Output, clrYellow+"==================== "+clrGreen+"GoMinewrap"+clrYellow+" ====================")
					fmt.Fprintln(color.Output, "     "+clrDarkMagenta+"GoMinewrap"+clrWhite+":"+clrEnd)
					fmt.Fprintln(color.Output, "     "+clrWhite+"- "+clrDarkYellow+"!help"+clrWhite+":  "+clrDarkCyan+"Display this help menu."+clrEnd)
					fmt.Fprintln(color.Output, "     "+clrWhite+"- "+clrDarkYellow+"!version"+clrWhite+":  "+clrDarkCyan+"Display the version of GoMinewrap"+clrEnd)
					fmt.Fprintln(color.Output, "     "+clrWhite+"- "+clrDarkYellow+"!filters"+clrWhite+":  "+clrDarkCyan+"Enable/disable the custom filters"+clrEnd)
					fmt.Fprintln(color.Output, "     "+clrWhite+"- "+clrDarkYellow+"!clear"+clrWhite+":  "+clrDarkCyan+"Clear the terminal."+clrEnd)
					fmt.Fprintln(color.Output, "     "+clrWhite+"- "+clrDarkYellow+"!server"+clrWhite+":  "+clrDarkCyan+"Manage all the servers."+clrEnd)
					fmt.Fprintln(color.Output, clrDarkCyan+"          Note: '*' will execute the command on every single server."+clrEnd)
					fmt.Fprintln(color.Output, "          "+clrWhite+"- "+clrDarkYellow+"!server [server]"+clrWhite+":  "+clrDarkCyan+"Switch to a different server."+clrEnd)
					fmt.Fprintln(color.Output, "          "+clrWhite+"- "+clrDarkYellow+"!server [server or *] stop"+clrWhite+":  "+clrDarkCyan+"Stop one or all the server(s)."+clrEnd)
					fmt.Fprintln(color.Output, "          "+clrWhite+"- "+clrDarkYellow+"!server [server or *] restart"+clrWhite+":  "+clrDarkCyan+"Restart one or all the server(s)."+clrEnd)
					fmt.Fprintln(color.Output, "          "+clrWhite+"- "+clrDarkYellow+"!server [server or *] backup"+clrWhite+":  "+clrDarkCyan+"Backup all the server files."+clrEnd)
					fmt.Fprintln(color.Output, "          "+clrWhite+"- "+clrDarkYellow+"!server [server or *] exec [command]"+clrWhite+":  "+clrEnd)
					fmt.Fprintln(color.Output, clrDarkCyan+"                Execute a command on any server without having to switching."+clrEnd)
					break cmd

				case "version":
					fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap v"+viper.GetString("version")+" by SuperPykkon."+clrEnd)
					break cmd

				case "filters":
					if len(strings.Fields(command)) == 2 {
						if strings.Fields(command)[1] == "on" {
							fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrDarkCyan+"Turned "+clrGreen+"on"+clrDarkCyan+" custom log filtering."+clrEnd)
							enableFilters = true
						} else if strings.Fields(command)[1] == "off" {
							fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrDarkCyan+"Turned "+clrRed+"off"+clrDarkCyan+" custom log filtering."+clrEnd)
							enableFilters = false
						} else {
							fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrDarkCyan+"Unknown arguement '"+strings.Fields(command)[1]+"'. !filters [on/off]"+clrEnd)
						}
					} else {
						fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrDarkCyan+"Usage: !filters [on/off]"+clrEnd)
					}
					break cmd

				case "server":
					if len(strings.Fields(command)) == 1 {
						fmt.Fprintln(color.Output, clrYellow+"Currently viewing the server '"+clrGreen+activeServer+clrYellow+"' 's console."+clrEnd)
						fmt.Fprintln(color.Output, clrYellow+"To switch to a different server, use "+clrMagenta+"!server [server]"+clrEnd)
						fmt.Fprintln(color.Output, clrYellow+"Available servers: "+clrEnd)
						for s, sd := range servers {
							if sd.status == IDLE || sd.status == FAILED {
								fmt.Fprintln(color.Output, clrWhite+" -  "+clrDarkCyan+s+"    "+clrRed+string(sd.status)+clrEnd)
							} else if sd.status == RUNNING {
								fmt.Fprintln(color.Output, clrWhite+" -  "+clrDarkCyan+s+"    "+clrGreen+string(sd.status)+clrEnd)
							}
						}
					} else if len(strings.Fields(command)) == 2 {
						for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
							if s == string(strings.Fields(command)[1]) {
								activeServer = strings.Fields(command)[1]
								serverCommandHandler("!clear")
								for i := 0; i < len(servers[activeServer].logs); i++ {
									fmt.Fprintln(color.Output, filters(servers[activeServer].logs[i].Log+"\n", "main"))
								}
								break cmd
							}
						}
						fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrRed+"Invalid server '"+string(strings.Fields(command)[1])+"'"+clrEnd)
					} else if len(strings.Fields(command)) >= 3 {
						if strings.Fields(command)[2] == "exec" {
							if len(strings.Fields(command)) > 3 {
								if strings.Fields(command)[1] == "*" {
									for _, sd := range servers {
										io.WriteString(sd.stdinPipe, strings.Join(strings.Fields(command)[3:], " ")+"\n")
									}
								} else {
									for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
										if s == string(strings.Fields(command)[1]) {
											io.WriteString(servers[string(strings.Fields(command)[1])].stdinPipe, strings.Join(strings.Fields(command)[3:], " ")+"\n")
											break cmd
										}
									}
									fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrRed+"Invalid server '"+string(strings.Fields(command)[1])+"'"+clrEnd)
								}
							}
						} else if strings.Fields(command)[2] == "stop" {
							serverCommandHandler("!server " + string(strings.Fields(command)[1]) + " exec stop")
						} else if strings.Fields(command)[2] == "restart" {
							if strings.Fields(command)[1] == "*" {

							} else {
								fmt.Println("!server " + string(strings.Fields(command)[1]) + " exec restart")
								serverCommandHandler("!server " + string(strings.Fields(command)[1]) + " exec restart")
								svr := servers[string(strings.Fields(command)[1])]
								svr.status = RESTARTING
								servers[string(strings.Fields(command)[1])] = svr
							}
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
								fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrGreen+"Begin backup of the server: "+clrYellow+"all servers"+clrGreen+"."+clrEnd)
								for name, _ := range viper.Get("server.servers").(map[string]interface{}) {
									fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrGreen+"  ...backing up the server: "+clrYellow+name+clrGreen+"."+clrEnd)
									if runtime.GOOS == "windows" {
										cmd := exec.Command("cmd", "/c", "robocopy", "../../"+viper.GetString("server.base")+viper.GetString("server.servers."+name+".dir"), viper.GetString("server.servers."+name+".dir"), "/MIR")
										cmd.Dir = viper.GetString("server.backup.dir") + t
										cmd.Run()
									} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
										cmd := exec.Command("cp", "-a", "../../"+viper.GetString("server.base")+viper.GetString("server.servers."+name+".dir")+".", viper.GetString("server.servers."+name+".dir"))
										cmd.Dir = viper.GetString("server.backup.dir") + t
										cmd.Run()
									}
								}
							} else {
								fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrGreen+"Begin backup of the server: "+clrYellow+string(strings.Fields(command)[1])+clrGreen+"."+clrEnd)
								if runtime.GOOS == "windows" {
									cmd := exec.Command("cmd", "/c", "robocopy", "../../"+viper.GetString("server.base")+viper.GetString("server.servers."+string(strings.Fields(command)[1])+".dir"), viper.GetString("server.servers."+string(strings.Fields(command)[1])+".dir"), "/MIR")
									cmd.Dir = viper.GetString("server.backup.dir") + t
									cmd.Run()
								} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
									cmd := exec.Command("cp", "-a", "../../"+viper.GetString("server.base")+viper.GetString("server.servers."+string(strings.Fields(command)[1])+".dir")+".", viper.GetString("server.servers."+string(strings.Fields(command)[1])+".dir"))
									cmd.Dir = viper.GetString("server.backup.dir") + t
									cmd.Run()
								}
							}
							fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrGreen+"Backup complete: "+clrYellow+viper.GetString("server.backup.dir")+t+"..."+clrEnd)
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
					fmt.Fprintln(color.Output, clrDarkCyan+"// "+clrYellow+"GoMinewrap > "+clrRed+"Unknown command, try '!help' for a list of commands."+clrEnd)
					break cmd
				}
			}
		}
	} else {
		switch strings.TrimSpace(command) {
		case "stop":
			fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"SERVER"+clrWhite+")"+clrDarkCyan+": "+clrMagenta+"Received stop command, exiting the server..."+clrEnd)
			io.WriteString(servers[activeServer].stdinPipe, command+"\n")
			servers[activeServer].process.Wait()

		case "restart":
			fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"SERVER"+clrWhite+")"+clrDarkCyan+": "+clrMagenta+"Received restart command, restarting the server..."+clrEnd)
			io.WriteString(servers[activeServer].stdinPipe, "stop\n")
			servers[activeServer].process.Wait()

		default:
			io.WriteString(servers[activeServer].stdinPipe, command+"\n")
			servers[activeServer].process.Wait()
		}
	}
}

func filters(text string, type_ string) string {
	line := strings.TrimSpace(text)
	var (
		logTime   bool
		logClrEnd bool
	)
	if type_ == "main" {
		if enableFilters {
			if regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).MatchString(line) {
				line = regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).ReplaceAllString(line, time.Now().Format("15:04:05"))
				logTime = true
			} else {
				logTime = false
			}

			if regexp.MustCompile(`INFO\]:`).MatchString(line) {
				if regexp.MustCompile(`\[/[0-9]+(?:\.[0-9]+){3}:[0-9]+\] logged in with entity id \d`).MatchString(line) {
					line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkGreen+"+"+clrDarkCyan+" INFO:"+clrGreen)
					logClrEnd = true
				} else if regexp.MustCompile(`[a-zA-Z0-9_.-] lost connection\:`).MatchString(line) {
					line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkRed+"-"+clrDarkCyan+" INFO:"+clrRed)
					logClrEnd = true
				} else {
					if regexp.MustCompile(`[a-zA-Z0-9_.-] issued server command\:`).MatchString(line) {
						line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrCyan+"|"+clrDarkCyan+" INFO:"+clrEnd)
						line = regexp.MustCompile(`\: `).ReplaceAllString(line, ": "+clrCyan)
						logClrEnd = true
					} else {
						line = regexp.MustCompile(`INFO\]:`).ReplaceAllString(line, clrDarkCyan+"| INFO:"+clrEnd)
					}
				}

				if logWarnSpacer {
					logWarnSpacer = false
				} else if logErrorSpacer {
					logErrorSpacer = false
				}
			} else if regexp.MustCompile(`WARN\]:`).MatchString(line) {
				line = regexp.MustCompile(`WARN\]:`).ReplaceAllString(line, clrDarkYellow+"! WARN:"+clrYellow)
				logWarnSpacer = true
				logClrEnd = true
			} else if regexp.MustCompile(`ERROR\]:`).MatchString(line) {
				line = regexp.MustCompile(`ERROR\]:`).ReplaceAllString(line, clrDarkRed+"x ERROR:"+clrRed)
				logErrorSpacer = true
				logClrEnd = true
			}

			line = regexp.MustCompile(`\[Server\]`).ReplaceAllString(line, clrMagenta+"[Server]"+clrEnd)

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
