package main

import (
	"net/http"
	"strconv"
	"time"
)

func init() {
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
