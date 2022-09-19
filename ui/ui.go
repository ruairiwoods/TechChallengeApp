// Copyright © 2022 Servian
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/servian/TechChallengeApp/db"
)

// Config configuration for ui package
type Config struct {
	Assets    http.FileSystem
	DB        db.Config
	Hostname  string
	Publicip  string
	Timestamp int64
}

// Start - start web server and handle web requets
func Start(cfg Config, listener net.Listener) {
	server := &http.Server{
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	mainRouter := mux.NewRouter()
	mainRouter.Handle("/healthcheck", healthcheckHandler(cfg))
	mainRouter.Handle("/healthcheck/", healthcheckHandler(cfg))

	mainRouter.Handle("/debug", debugHandler(cfg))

	apiRouter := mainRouter.PathPrefix("/api").Subrouter()
	apiHandler(cfg, apiRouter)

	uiRouter := mainRouter.PathPrefix("/").Subrouter()
	uiHandler(cfg, uiRouter)

	http.Handle("/", mainRouter)
	go server.Serve(listener)
}

func healthcheckHandler(cfg Config) http.Handler {
	db := db.GetDatabase(cfg.DB)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := db.GetAllTasks(cfg.DB)

		if err != nil {
			fmt.Fprintf(w, "Error: db connection down")
			return
		}

		fmt.Fprintf(w, "OK")
	})
}

func debugHandler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.DB.DbUser = "********"
		cfg.DB.DbPassword = "********"
		cfg.Publicip = publicIP()
		cfg.Hostname, _ = os.Hostname()
		cfg.Timestamp = time.Now().Unix()
		bytes, err := json.Marshal(cfg)

		if err != nil {
			fmt.Fprintf(w, "{\"error\": \"JSON Marshal error\"}")
			return
		}
		fmt.Fprintf(w, string(bytes))
	})
}

func publicIP() string {
	url := "https://checkip.amazonaws.com/"
	resp, err := http.Get(url)
	if err != nil {
		return "connect-error"
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "read-error"
	}
	return string(ip)
}
