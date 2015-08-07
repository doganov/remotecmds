package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

var commands commandTable
var events chan event
var ids chan int
var statuses chan statusTable

type commandTable []*command

func (t commandTable) Len() int {
	return len(t)
}

func (t commandTable) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t commandTable) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}

type event struct {
	id int
	c *command
	t eventType
	ocurred time.Time
}

func (e event) log() {
	log.Printf("[%d] %s %s", e.id, e.c.Name, e.t)
}

type eventType int

const (
	eventTypeBegin = iota
	eventTypeEnd
)

func (t eventType) String() string {
	switch t {
	case eventTypeBegin: return "begin"
	case eventTypeEnd: return "end"
	default: return "UNKNOWN"
	}
}

type status struct {
	id int
	c *command
	started time.Time
}

type statusTable []status

func (t statusTable) removeIndex(i int) statusTable {
	return append(t[:i], t[i+1:]...)
}

func (t statusTable) removeId(id int) statusTable {
	// find index
	for i, row := range(t) {
		if row.id == id {
			return t.removeIndex(i)
		}
	}
	panic(fmt.Sprintf("Unknown id: %d", id))
}

func (t statusTable) clone() statusTable {
	t2 := make([]status, len(t))
	copy(t2, t)
	return statusTable(t2)
}

func main() {
	events = make(chan event)
	ids = make(chan int)
	statuses = make(chan statusTable)
	
	go generateIds()
	go manageStatuses()
	
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe("localhost:4000", nil))
}

func generateIds() {
	id := 1
	for {
		ids <- id
		id++
	}
}

func manageStatuses() {
	table := make(statusTable, 0)

	for {
		select {
		case event := <-events:
			if event.t == eventTypeBegin {
				table = append(table,
					status{event.id, event.c, event.ocurred})
			} else {
				table = table.removeId(event.id)
			}
		case statuses <- table.clone():
		}
	}
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func sendEvent(id int, c *command, t eventType) {
	e := event{id, c, t, time.Now().UTC()}
	e.log()
	events <- e
}

func wrap(c *command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			httpError(w, http.StatusMethodNotAllowed)
			return
		}

		id := <-ids
		sendEvent(id, c, eventTypeBegin)
		defer func() { sendEvent(id, c, eventTypeEnd) }()

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
	sort.Sort(commands)

	handler := wrap(c)
	http.Handle(c.Name, handler)
	if c.Name == "/help" {
		http.Handle("/", handler)
	}
}

func init() {
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
		"/status",
		"List currently running requests",
		func(w http.ResponseWriter, r *http.Request) {
			table := <-statuses
			now := time.Now().UTC()
			fmt.Fprintln(w, "No\tId\tDur (ms)\tCommand")
			fmt.Fprintln(w, "--\t--\t--------\t-------")
			for i, row := range(table) {
				duration := int64(now.Sub(row.started))
				duration = duration / int64(time.Millisecond) // truncate
				fmt.Fprintf(w, "%d\t%d\t%8d\t%s\n",
					i+1, row.id, duration, row.c.Name)
			}
		},
	})
}
