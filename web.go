package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

func webconHandler() {
	// TODO
	http.Handle("/", webconAuthValidate(webconRootHandler))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/login", webconAuthLogin)
	http.HandleFunc("/logout", webconAuthValidate(webconAuthLogout))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("static/web/html/resources"))))

	http.ListenAndServe(viper.GetString("web.host")+":"+viper.GetString("web.port"), nil)
}

func webconRootHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(MyKey).(Claims)
	if !ok {
		http.Redirect(w, r, "/login", 307)
	}

	if _, err := os.Stat(viper.GetString("web.dir") + "index.html"); os.IsNotExist(err) {
		fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing file 'index.html' or invalid path."+clrEnd)
		fmt.Fprintln(w, "An internal error has occured.")
	} else {
		fmt.Fprintln(w, "TODO Do dashboard")

	}
}

func webconAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if _, err := os.Stat(viper.GetString("web.dir") + "login.html"); os.IsNotExist(err) {
			fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEB"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing file 'login.html' or invalid path."+clrEnd)
			fmt.Fprintln(w, "An internal error has occured.")
		} else {
			t := template.Must(template.ParseFiles(viper.GetString("web.dir") + "login.html"))
			temp := consoleLogin{Status: ""}
			t.Execute(w, temp)
		}
	} else {
		r.ParseForm()
		// Get sha256 from the form password
		pass256 := sha256.New()
		io.WriteString(pass256, r.Form["password"][0])
		if users[r.Form["username"][0]] != "" && users[r.Form["username"][0]] == fmt.Sprintf("%x", pass256.Sum(nil)) {

			if viper.GetBool("web.messages.login_success") {
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEB"+clrWhite+")"+clrDarkCyan+": "+clrGreen+"Authentication successful from: "+clrYellow+strings.Split(r.RemoteAddr, ":")[0]+clrGreen+"\n                           user: "+clrYellow+r.Form["username"][0]+clrEnd)
			}
			// Expires the token and cookie in 2 hour
			expireToken := time.Now().Add(time.Hour * 2).Unix()
			expireCookie := time.Now().Add(time.Hour * 2)

			// We'll manually assign the claims but in production you'd insert values from a database
			claims := Claims{
				r.Form["username"][0],
				jwt.StandardClaims{
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
			if _, err := os.Stat(viper.GetString("web.dir") + "login.html"); os.IsNotExist(err) {
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing file 'login.html' or invalid path."+clrEnd)
				fmt.Fprintln(w, "An internal error has occured.")
			} else {
				if viper.GetBool("web.messages.login_fail") {
					fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Authentication failed from: "+clrYellow+strings.Split(r.RemoteAddr, ":")[0]+clrRed+"\n                           reason: Invalid username/password."+clrEnd)
				}
				t := template.Must(template.ParseFiles(viper.GetString("web.dir") + "login.html"))
				temp := consoleLogin{Status: "Invalid username or password."}
				t.Execute(w, temp)
			}
		}
	}
}

// middleware to protect private pages
func webconAuthValidate(page http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		// If no Auth cookie is set then return a 404 not found
		cookie, err := req.Cookie("Auth")
		if err != nil {
			http.Redirect(res, req, "/login", 307)
			return
		}

		// Return a Token using the cookie
		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

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
	if viper.GetBool("web.messages.ws_connect") {
		fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+") "+clrMagenta+"WS"+clrDarkCyan+": "+clrGreen+"Connection established from: "+clrYellow+IP+clrDarkCyan+" ["+strconv.Itoa(len(wsConns))+"]"+clrEnd)
	}
	for {
		exec := wsExec{}
		err := conn.ReadJSON(&exec)
		if err != nil {
			delete(wsConns, CID)
			if viper.GetBool("web.messages.ws_disconnect") {
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+") "+clrMagenta+"WS"+clrDarkCyan+": "+clrRed+"Connection terminated from: "+clrYellow+IP+clrDarkCyan+" ["+strconv.Itoa(len(wsConns))+"]"+clrEnd)
			}
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
				for i := 0; i <= len(servers[server].logs); i++ {
					if servers[server].logs[i].Type == "server" {
						conn.WriteJSON(filters(servers[server].logs[i].Log))
					}
				}
			} else {
				if server == activeServer {
					if viper.GetBool("web.messages.exec_command") {
						fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrYellow+claims.Username+clrEnd+" executed a server command: "+clrCyan+"/"+exec.Command+clrEnd)
					}
				}
				if viper.GetBool("web.messages.exec_command") {
					servers[server].logs[len(servers[server].logs)] = slogs{Type: "web", Log: time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrYellow + claims.Username + clrEnd + " executed a server command: " + clrCyan + "/" + exec.Command + clrEnd}
				}

				time.Sleep(time.Millisecond)
				io.WriteString(servers[server].stdinPipe, exec.Command+"\n")
			}
		}
	}
}

func filters(text string) string {
	line := strings.TrimSpace(text)
	var (
		logTime   bool
		logClrEnd bool
	)

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
