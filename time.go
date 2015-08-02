package main

import (
	"fmt"
	"net/http"
	"time"
)

func init() {
	define(&command{
		"/time",
		"Current UTC time",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, time.Now().UTC())
		},
	})
}

