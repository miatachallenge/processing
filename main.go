package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sdomino/scribble"
	r "gopkg.in/gorethink/gorethink.v4"
)

var (
	bind    = flag.String("bind", ":8070", "bind address")
	rdbAddr = flag.String("rdb_addr", "127.0.0.1:28015", "rdb address")
	dbPath  = flag.String("db_path", "./_database", "db dir path")
)

type stateStruct struct {
	Database struct {
		Count int `json:"count"`
	} `json:"database"`
	Params struct {
		Autostart         string `json:"autostart"`
		Tag               string `json:"tag"`
		Convert           string `json:"convert"`
		Validate          string `json:"validate"`
		Upstream          string `json:"upstream"`
		LegacySpreadsheet string `json:"legacy_spreadsheet"`
		LegacyAddress     string `json:"legacy_address"`
	} `json:"params"`
	Upstream struct {
		Processing bool `json:"processing"`
		Count      int  `json:"count"`
	} `json:"upstream"`
	LegacySpreadsheet struct {
		Count int `json:"count"`
	} `json:"legacy_spreadsheet"`
	Tags       tagsSlice `json:"tags"`
	RecentLaps []*Lap    `json:"recent_laps"`
}

var (
	stateLock sync.RWMutex
	state     = stateStruct{
		Tags: tagsSlice{},
	}

	db  *r.Session
	fdb *scribble.Driver
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	flag.Parse()

	// Connect to the databases
	var err error
	db, err = r.Connect(r.ConnectOpts{
		Address:  *rdbAddr,
		Database: "mctracker",
	})
	if err != nil {
		panic(err)
	}
	fdb, err = scribble.New(*dbPath, nil)
	if err != nil {
		panic(err)
	}

	// Load from the local database
	loadDefaultParams()

	// Load tags from RethinkDB
	loadTags()

	// And load the spreadsheet
	syncSpreadsheet()

	// Then start the processing engine
	go func() {
		time.Sleep(time.Second) // just in case
		stateLock.RLock()
		autostartEnabled := state.Params.Autostart == "true"
		currentlyRunning := state.Upstream.Processing
		stateLock.RUnlock()

		if autostartEnabled && !currentlyRunning {
			startProcessing()
		}
	}()

	/*
		vm := otto.New()
		vm.Run(`function process(input) {
			return parseInt(input.toString().substr(6))
		}`)

		value, err := vm.Call("process", nil, "0x1C010063")
		if err != nil {
			panic(err)
		}
		x, err := value.ToInteger()
		if err != nil {
			panic(err)
		}
		fmt.Println("id", x)

		vm.Run(`function validate(time) {
			return time > 30 && time < 300
		}`)
	*/

	// WS handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		kill := make(chan struct{})
		defer func() {
			kill <- struct{}{}
		}()

		stateLock.RLock()
		err = conn.WriteJSON(state)
		stateLock.RUnlock()
		if err != nil {
			log.Printf("WS died: %s", err)
			return
		}

		go func() {
			defer close(kill)

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					stateLock.RLock()
					err := conn.WriteJSON(state)
					stateLock.RUnlock()
					if err != nil {
						log.Printf("WS died: %s", err)
						return
					}
				case <-kill:
					return
				}
			}
		}()

		for {
			var input struct {
				Type  string `json:"type"`
				Param string `json:"param"`
				Value string `json:"value"`
			}
			if err := conn.ReadJSON(&input); err != nil {
				log.Printf("WS died: %s", err)
				return
			}

			if input.Type == "update" {
				setParam(input.Param, input.Value)
				continue
			}

			if input.Type == "processing" {
				if input.Value == "true" {
					stateLock.RLock()
					currentlyProcessing := state.Upstream.Processing
					stateLock.RUnlock()

					if !currentlyProcessing {
						go startProcessing()
					}
				} else if input.Value == "false" {
					stateLock.RLock()
					currentlyProcessing := state.Upstream.Processing
					stateLock.RUnlock()

					if currentlyProcessing {
						go stopProcessing()
					}
				}
			}
		}
	})

	http.Handle("/", http.FileServer(http.Dir("frontend")))
	http.ListenAndServe(*bind, nil)
}
