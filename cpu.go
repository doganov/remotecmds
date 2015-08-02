package main

import (
	"fmt"
	"net/http"

	//"golang.org/x/sys/unix"
)

func init() {
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
}
