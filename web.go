package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

func webconHandler() {
	time.Sleep(time.Millisecond * 1500)

	http.Handle("/", webconAuthValidate(webconRootHandler))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/login", webconAuthLogin)
	http.HandleFunc("/logout", webconAuthValidate(webconAuthLogout))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("static/web/html/resources"))))

	http.ListenAndServe(viper.GetString("webcon.host")+":"+viper.GetString("webcon.port"), nil)
}

func webconRootHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(MyKey).(Claims)
	if !ok {
		http.Redirect(w, r, "/login", 307)
	}

	if _, err := os.Stat(viper.GetString("webcon.dir") + "index.html"); os.IsNotExist(err) {
		fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing file 'index.html' or invalid path."+clrEnd)
		fmt.Fprintln(w, "An internal error has occured.")
	} else {
		var blocked bool
		var server bool = false
		for _, a := range viper.Get("webcon.blacklist.IP").([]interface{}) {
			if strings.Split(r.RemoteAddr, ":")[0] == a.(string) {
				fmt.Fprintln(w, "You do not have permission to access this page.")
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Terminated connection from blacklisted IP: "+clrYellow+a.(string)+clrEnd)
				blocked = true
			}
		}

		for _, a := range viper.Get("webcon.blacklist.users").([]interface{}) {
			if claims.Username == a.(string) {
				fmt.Fprintln(w, "You do not have permission to access this page.")
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Terminated connection from blacklisted user: "+clrYellow+a.(string)+clrEnd)
				blocked = true
			}
		}

		for s, _ := range viper.Get("server.servers").(map[string]interface{}) {
			if s == strings.Join(r.URL.Query()["server"], " ") {
				server = true
			}
		}

		if !blocked {
			if server {
				cookie, _ := r.Cookie("Auth")
				var temp consoleTemplate
				if len(r.URL.Query()["server"]) > 0 {
					temp = consoleTemplate{Username: claims.Username, Token: cookie.Value, Server: strings.Join(r.URL.Query()["server"], " ")}
				} else {
					temp = consoleTemplate{Username: claims.Username, Token: cookie.Value, Server: viper.GetString("server.primary")}
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
			fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing file 'login.html' or invalid path."+clrEnd)
			fmt.Fprintln(w, "An internal error has occured.")
		} else {
			var blocked bool = false
			for _, a := range viper.Get("webcon.blacklist.IP").([]interface{}) {
				if strings.Split(r.RemoteAddr, ":")[0] == a.(string) {
					fmt.Fprintln(w, "You do not have permission to access this page.")
					fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Terminated connection from blacklisted IP: "+clrYellow+a.(string)+clrEnd)
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
			if viper.GetBool("webcon.messages.login_success") {
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrGreen+"Authentication successful from: "+clrYellow+strings.Split(r.RemoteAddr, ":")[0]+clrGreen+"\n                           user: "+clrYellow+r.Form["username"][0]+clrEnd)
			}
			// Expires the token and cookie in 1 hour
			expireToken := time.Now().Add(time.Hour * 1).Unix()
			expireCookie := time.Now().Add(time.Hour * 1)

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
			if _, err := os.Stat(viper.GetString("webcon.dir") + "login.html"); os.IsNotExist(err) {
				fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Error: missing file 'login.html' or invalid path."+clrEnd)
				fmt.Fprintln(w, "An internal error has occured.")
			} else {
				if viper.GetBool("webcon.messages.login_fail") {
					fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrRed+"Authentication failed from: "+clrYellow+strings.Split(r.RemoteAddr, ":")[0]+clrRed+"\n                           reason: Invalid username/password."+clrEnd)
				}
				t := template.Must(template.ParseFiles(viper.GetString("webcon.dir") + "login.html"))
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
	if viper.GetBool("webcon.messages.ws_connect") {
		fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+") "+clrMagenta+"WS"+clrDarkCyan+": "+clrGreen+"Connection established from: "+clrYellow+IP+clrDarkCyan+" ["+strconv.Itoa(len(wsConns))+"]"+clrEnd)
	}
	for {
		exec := wsExec{}
		err := conn.ReadJSON(&exec)
		if err != nil {
			delete(wsConns, CID)
			if viper.GetBool("webcon.messages.ws_disconnect") {
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
						conn.WriteJSON(filters(servers[server].logs[i].Log, "webcon"))
					}
				}
			} else {
				if server == activeServer {
					if viper.GetBool("webcon.messages.exec_command") {
						fmt.Fprintln(color.Output, time.Now().Format("15:04:05")+clrDarkCyan+" | INFO: "+clrWhite+"("+clrDarkMagenta+"WEBCON"+clrWhite+")"+clrDarkCyan+": "+clrYellow+claims.Username+clrEnd+" executed a server command: "+clrCyan+"/"+exec.Command+clrEnd)
					}
				}
				if viper.GetBool("webcon.messages.exec_command") {
					servers[server].logs[len(servers[server].logs)] = slogs{Type: "webcon", Log: time.Now().Format("15:04:05") + clrDarkCyan + " | INFO: " + clrWhite + "(" + clrDarkMagenta + "WEBCON" + clrWhite + ")" + clrDarkCyan + ": " + clrYellow + claims.Username + clrEnd + " executed a server command: " + clrCyan + "/" + exec.Command + clrEnd}
				}

				time.Sleep(time.Millisecond)
				io.WriteString(servers[server].stdinPipe, exec.Command+"\n")
			}
		}
	}
}
