package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

func init() {
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
}
