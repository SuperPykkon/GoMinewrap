package main

import (
	"bufio"
	"io"
	"os/exec"
)

// TODO Make server control functions to control the server here

func serverLogHandler(name string, process *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, status string) {
	output := bufio.NewScanner(stdout)
	go func() {
		for output.Scan() {
			if output.Text() != "" {
				servers[name].logs[len(servers[name].logs)] = slogs{Type: "server", Log: output.Text()}
				for _, ud := range wsConns {
					if ud.Server == name {
						ud.Conn.WriteJSON(filters(output.Text()))
					}
				}
			}
		}
	}()
}
