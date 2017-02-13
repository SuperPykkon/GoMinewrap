package main

import (
    "fmt"
    "strings"
    "os/exec"
    "os"
    "bufio"
    "regexp"
    "time"

    "github.com/fatih/color"
)

const runCmd = "java -Xms512M -Xmx512M -XX:+UseConcMarkSweepGC -jar spigot.jar"

const (
    clrRed = "\x1b[31m"
    clrGreen = "\x1b[32m"
    clrYellow = "\x1b[33m"
    clrBlue = "\x1b[34m"
    clrMagenta = "\x1b[35m"
    clrCyan = "\x1b[36m"
    clrWhite = "\x1b[37m"
    clrEnd = "\x1b[0m"
)

var (
    logWarnSpacer bool
    logErrorSpacer  bool
)

func main() {
    fmt.Fprintln(color.Output, time.Now().Format("15:04:05") + clrCyan + " | INFO " + clrWhite + "(" + clrMagenta + "SERVER" + clrWhite + ")" + clrCyan + ": " + clrWhite + "Starting the server...")
    process := exec.Command(strings.Fields(runCmd)[0], strings.Fields(runCmd)[1:]...)
    process.Stdin = os.Stdin

    stdout, err := process.StdoutPipe()
  	if err != nil { fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for the process:\n" + err.Error()) }

  	scanner := bufio.NewScanner(stdout)
    go func() {
        for scanner.Scan() {
    			  if scanner.Text() != "" {
                filters(scanner.Text() + "\n")
            }
    		}
    }()

    if err = process.Start(); err != nil {
        fmt.Println(err)
    }

    if err = process.Wait(); err != nil {
        fmt.Println(err)
    }
}

func filters(line string) {
    text := strings.TrimSpace(line)
    var (
        logTime bool
        logClrEnd bool
    )

    if regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).MatchString(text) {
        text = regexp.MustCompile(`\[(\d{2}):(\d{2}):(\d{2})`).ReplaceAllString(text, time.Now().Format("15:04:05"))
        logTime = true
    } else { logTime = false }

    if regexp.MustCompile(`INFO\]:`).MatchString(text) {
        if regexp.MustCompile(`\[/[0-9]+(?:\.[0-9]+){3}:[0-9]+\] logged in with entity id \d`).MatchString(text) {
            text = regexp.MustCompile(`INFO\]:`).ReplaceAllString(text, clrGreen + "+" + clrCyan + " INFO:" + clrGreen)
            logClrEnd = true
        } else if regexp.MustCompile(`[a-zA-Z0-9_.-] lost connection\:`).MatchString(text) {
            text = regexp.MustCompile(`INFO\]:`).ReplaceAllString(text, clrRed + "-" + clrCyan + " INFO:" + clrRed)
            logClrEnd = true
        } else {
            text = regexp.MustCompile(`INFO\]:`).ReplaceAllString(text, clrCyan + "| INFO:" + clrEnd)
        }

        if logWarnSpacer {
            logWarnSpacer = false
        } else if logErrorSpacer {
            logErrorSpacer = false
        }
    } else if regexp.MustCompile(`WARN\]:`).MatchString(text) {
        text = regexp.MustCompile(`WARN\]:`).ReplaceAllString(text, clrYellow + "! WARN:")
        logWarnSpacer = true
        logClrEnd = true
    } else if regexp.MustCompile(`ERROR\]:`).MatchString(text) {
        text = regexp.MustCompile(`ERROR\]:`).ReplaceAllString(text, clrRed + "x ERROR:")
        logErrorSpacer = true
        logClrEnd = true
    }

    text = regexp.MustCompile(`\[Server\]`).ReplaceAllString(text, clrMagenta + "[Server]" + clrEnd)

    if logWarnSpacer && !logTime {
        fmt.Fprintln(color.Output, "              " + clrYellow + text + clrEnd)
    } else if logErrorSpacer && !logTime {
        fmt.Fprintln(color.Output, "              " + clrRed + text + clrEnd)
    } else if logClrEnd {
        fmt.Fprintln(color.Output, text + clrEnd)
    } else {
        fmt.Fprintln(color.Output, text)
    }
}
