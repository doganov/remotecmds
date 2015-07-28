package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	//"golang.org/x/sys/unix"
)

var commands []*command
var updates chan statusUpdate

type statusUpdate struct {
	c *command
	started bool
}

func main() {
	updates = make(chan statusUpdate)
	
	defineCommands()
	go manageStatuses()
	
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe("localhost:4000", nil))
}

func manageStatuses() {
	for {
		select {
		case update := <-updates:
			log.Println(update.c.Name, update.started)
		}
	}
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func wrap(c *command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			httpError(w, http.StatusMethodNotAllowed)
			return
		}

		updates <- statusUpdate{c, true}
		defer func() { updates <- statusUpdate{c, false} }()

		c.Func(w, r)
	}
}

type command struct {
	Name string
	Doc string
	Func http.HandlerFunc
}

func define(c *command) {
	commands = append(commands, c)

	handler := wrap(c)
	http.Handle(c.Name, handler)
	if c.Name == "/help" {
		http.Handle("/", handler)
	}
}

func defineCommands() {
	define(&command{
		"/help",
		"Returns this text",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Available commands:\n")
			for _, cmd := range(commands) {
				fmt.Fprintf(w, "%s\t%s\n", cmd.Name, cmd.Doc)
			}
		},
	})

	define(&command{
		"/time",
		"Current UTC time",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, time.Now().UTC())
		},
	})

	define(&command{
		"/cpu",
		"Current CPU usage",
		func(w http.ResponseWriter, r *http.Request) {
			//sysinfo := &unix.Sysinfo_t{};
			//err := unix.Sysinfo(sysinfo)
			//if err != nil {
			//	log.Println(err)
			//	http.Error(w, err.String(), 500)
			//}
			fmt.Fprintln(w, "5")
		},
	})

	define(&command{
		"/say",
		"Make computer \"say\" something (passed with v parameter)",
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				httpError(w, http.StatusBadRequest)
				return
			}

			v := r.Form.Get("v")
			fmt.Fprintln(w, v)
			
			cmd := exec.Command("say")
			cmd.Stdin = strings.NewReader(v)
			cmd.Stdout = new(bytes.Buffer)
			err = cmd.Run()
			if err != nil {
				fmt.Fprintln(w, err)
				httpError(w, http.StatusInternalServerError)
			}
		},
	})

	define(&command{
		"/sleep",
		"Sleep s number of seconds",
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				httpError(w, http.StatusBadRequest)
				return
			}

			sStr := r.Form.Get("s")
			s, err := strconv.ParseUint(sStr, 0, 64)
			if err != nil {
				httpError(w, http.StatusBadRequest)
				return
			}

			time.Sleep(time.Duration(s) * time.Second)			
		},
	})
}
