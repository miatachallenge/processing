package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flimzy/kivik"
	_ "github.com/go-kivik/couchdb" // the CouchDB driver
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	r "gopkg.in/gorethink/gorethink.v4"
)

type Lap struct {
	ID        int64 `json:"id"`
	Driver    int   `json:"driver"`
	Timestamp int64 `json:"timestamp"`
	LapTime   int64 `json:"lap_time"`
}

type Event struct {
	ID        int64  `json:"id" db:"id" gorethink:"internal_id"`
	Key       string `gorethink:"key"`
	Name      string `json:"-" db:"name" gorethink:"name"`
	TagID     string `json:"tag_id" db:"tag_id" gorethink:"tag_id"`
	Timestamp int64  `json:"timestamp" db:"timestamp" gorethink:"timestamp"`
	Antenna   int    `json:"antenna" db:"antenna" gorethink:"antenna"`
}

type LegacyLap struct {
	CouchID      string  `json:"_id,omitempty"`
	LapID        int64   `json:"id"`
	DriverID     int     `json:"driverId"`
	Microtime    int64   `json:"microtime"`
	CreatedAt    string  `json:"createdAt"`
	Correction   int64   `json:"correction"`
	Verification *string `json:"verification"`
}

const (
	recentLapsCount = 10
	designDocs      = 0
)

var (
	lastPingMap      = map[string]int64{}
	lastPingMutex    sync.Mutex
	cancelProcessing context.CancelFunc
	cancelMutex      sync.Mutex
)

func process(ctx context.Context) error {
	stateLock.RLock()
	processType := state.Params.Upstream
	stateLock.RUnlock()

	if processType != "legacy_couchdb" {
		return errors.New("invalid upstream type")
	}

	// We're spawning three sub-goroutines
	// 1) A counter updater that just displays the count of items in the target db
	// 2) A RethinkDB-backed pipeline which figures out laps and validates them
	// 3) An upstream saver that writes the laps to the database
	var (
		counterError  = make(chan error)
		sourceError   = make(chan error)
		upstreamError = make(chan error)
	)

	stateLock.RLock()
	separatorIndex := strings.LastIndex(state.Params.LegacyAddress, "/")
	var (
		couchAddress  = state.Params.LegacyAddress[:separatorIndex+1]
		couchDatabase = state.Params.LegacyAddress[separatorIndex+1:]
	)
	stateLock.RUnlock()

	// So 1 - the counter
	go func() {
		client, err := kivik.New(ctx, "couch", couchAddress)
		if err != nil {
			counterError <- errors.Wrap(err, "unable to connect to couchdb")
			return
		}
		db, err := client.DB(ctx, couchDatabase)
		if err != nil {
			counterError <- errors.Wrap(err, "unable to select the database")
			return
		}

		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

	mainLoop:
		for {
			select {
			case <-ticker.C:
				stats, err := db.Stats(ctx)
				if err != nil {
					log.Printf("processing: unable to get db stats: %s", err)
					break mainLoop
				}

				stateLock.Lock()
				count := int(stats.DocCount) - designDocs
				log.Printf("Updated the upstream count: %d", count)
				state.Upstream.Count = count
				stateLock.Unlock()
			case <-ctx.Done():
				break mainLoop
			}
		}

		counterError <- context.Canceled
		return
	}()

	upstreamQueue := make(chan *Lap, 128) // welp

	// 2 - the RethinkDB processing
	go func() {
		stateLock.RLock()
		tag := state.Params.Tag
		convertCode := state.Params.Convert
		validateCode := state.Params.Validate
		stateLock.RUnlock()

		defer func() {
			stateLock.Lock()
			state.Database.Count = 0
			stateLock.Unlock()

			lastPingMutex.Lock()
			lastPingMap = map[string]int64{}
			lastPingMutex.Unlock()
		}()

		// We're running Otto here for custom scripts
		vm := otto.New()
		if _, err := vm.Run(convertCode); err != nil {
			sourceError <- errors.Wrap(err, "convert code error")
			return
		}
		if _, err := vm.Run(validateCode); err != nil {
			sourceError <- errors.Wrap(err, "convert code error")
			return
		}

		processEvent := func(event *Event) error {
			defer func() {
				stateLock.Lock()
				state.Database.Count++
				stateLock.Unlock()
			}()

			if event.Timestamp == 0 {
				return nil
			}

			// Do the conversion on the tag ID
			conversionResult, err := vm.Call("convert", nil, event.TagID)
			if err != nil {
				return errors.Wrapf(err, "convert code execution error - %s", event.TagID)
			}

			conversionResultString, err := conversionResult.ToString()
			if err != nil {
				return errors.Wrapf(err, "convert code result error - %s", event.TagID)
			}

			if err := processReadout(upstreamQueue, vm, event.ID, conversionResultString, event.Timestamp); err != nil {
				return errors.Wrapf(err, "processing error - %s", event.TagID)
			}

			return nil
		}

		// Initial load
		cursor, err := r.Table("records").
			GetAllByIndex("key", tag).
			OrderBy("timestamp").
			Run(db, r.RunOpts{
				Context: ctx,
			})
		if err != nil {
			sourceError <- errors.Wrap(err, "unable to query for previous records")
			return
		}

		for {
			event := &Event{}
			if !cursor.Next(event) {
				break
			}

			if err := processEvent(event); err != nil {
				sourceError <- err
				return
			}
		}
		if err := cursor.Err(); err != nil {
			sourceError <- errors.Wrap(err, "unable to fetch the existing fields")
			return
		}
		if err := cursor.Close(); err != nil {
			sourceError <- errors.Wrap(err, "unable to close the existing fields cursor")
			return
		}

		// Changefeed
		cursor, err = r.Table("records").
			GetAllByIndex("key", tag).
			Changes().
			Run(db, r.RunOpts{
				Context: ctx,
			})
		if err != nil {
			sourceError <- errors.Wrap(err, "unable to query for records changes")
			return
		}

		for {
			var change struct {
				OldVal *Event `gorethink:"old_val"`
				NewVal *Event `gorethink:"new_val"`
			}
			if !cursor.Next(&change) {
				break
			}
			if change.NewVal == nil {
				continue
			}

			if err := processEvent(change.NewVal); err != nil {
				sourceError <- err
				return
			}
		}
		if err := cursor.Err(); err != nil {
			sourceError <- errors.Wrap(err, "unable to fetch the records changes")
			return
		}
		if err := cursor.Close(); err != nil {
			sourceError <- errors.Wrap(err, "unable to close the records changes cursor")
			return
		}

		sourceError <- context.Canceled
		return
	}()

	// 3 - the upstream
	go func() {
		client, err := kivik.New(ctx, "couch", couchAddress)
		if err != nil {
			counterError <- errors.Wrap(err, "unable to connect to couchdb")
			return
		}
		db, err := client.DB(ctx, couchDatabase)
		if err != nil {
			counterError <- errors.Wrap(err, "unable to select the database")
			return
		}

		for {
			lap := <-upstreamQueue
			log.Printf("Uploading a new lap: %+v", lap)

			stateLock.Lock()
			state.Upstream.Count++
			state.RecentLaps = append(state.RecentLaps, lap)
			if len(state.RecentLaps) > recentLapsCount {
				state.RecentLaps = state.RecentLaps[len(state.RecentLaps)-recentLapsCount:]
			}
			stateLock.Unlock()

			idString := strconv.FormatInt(lap.ID, 10)
			if _, err := db.Get(ctx, idString); err != nil {
				if kivik.Reason(err) != "Not Found: missing" {
					counterError <- errors.Wrapf(err, "unable to decide on lap %d", lap.ID)
					return
				}

				docID, _, err := db.CreateDoc(ctx, LegacyLap{
					CouchID:   strconv.FormatInt(lap.ID, 10),
					LapID:     lap.ID,
					DriverID:  lap.Driver,
					Microtime: lap.LapTime * 1000, // milliseconds to microseconds
					CreatedAt: time.Unix(0, lap.Timestamp*int64(time.Millisecond)). // milliseconds to nanoseconds
													Format("2006-01-02T15:04:05.999Z"),
					Correction:   0,
					Verification: nil,
				})
				if err != nil {
					counterError <- errors.Wrapf(err, "unable to insert lap %d", lap.ID)
					return
				}

				log.Printf("Lap %d by %d (%d) inserted as %s", lap.ID, lap.Driver, lap.LapTime, docID)
			} else {
				log.Printf("Lap %d already exists.", lap.ID)
			}
		}
	}()

	// The error handlers:
	go func() {
		err := <-counterError
		if err == nil {
			return
		}
		log.Printf("counter failed: %s", err)

		cancelMutex.Lock()
		if cancelProcessing != nil {
			cancelProcessing()
		}
		cancelMutex.Unlock()
	}()

	go func() {
		err := <-sourceError
		if err == nil {
			return
		}
		log.Printf("source failed: %s", err)

		cancelMutex.Lock()
		if cancelProcessing != nil {
			cancelProcessing()
		}
		cancelMutex.Unlock()
	}()

	go func() {
		err := <-upstreamError
		if err == nil {
			return
		}
		log.Printf("upstream failed: %s", err)

		cancelMutex.Lock()
		if cancelProcessing != nil {
			cancelProcessing()
		}
		cancelMutex.Unlock()
	}()

blockLoop:
	for {
		select {
		case <-ctx.Done():
			break blockLoop
		}
	}

	return nil
}

func processReadout(
	result chan *Lap, vm *otto.Otto,
	id int64, tag string, timestamp int64,
) error {
	lastPingMutex.Lock()
	existing, ok := lastPingMap[tag]
	lastPingMutex.Unlock()

	if ok {
		if existing > timestamp {
			log.Printf("Conflict: more recent event newer than the existing one - %s", tag)
		} else {
			lapTime := timestamp - existing

			lastPingMutex.Lock()
			lastPingMap[tag] = timestamp
			lastPingMutex.Unlock()

			// Do the conversion on the tag ID
			validationResult, err := vm.Call("validate", nil, lapTime)
			if err != nil {
				return errors.Wrapf(err, "validate code execution error - %s/%d", tag, timestamp)
			}
			resultBool, err := validationResult.ToBoolean()
			if err != nil {
				return errors.Wrapf(err, "validate code result error - %s/%d", tag, timestamp)
			}

			if !resultBool {
				return nil
			}

			driversLock.RLock()
			driver, ok := driversMap[tag]
			driversLock.RUnlock()
			if !ok {
				log.Printf("driver for tag %s not found", tag)
				return nil
			}

			result <- &Lap{
				ID:        id,
				Driver:    driver,
				Timestamp: timestamp,
				LapTime:   lapTime,
			}
		}
	} else {
		// First time seeing it
		lastPingMutex.Lock()
		lastPingMap[tag] = timestamp
		lastPingMutex.Unlock()

		log.Printf("First time seeing an event for the tag %s", tag)
	}

	return nil
}

func startProcessing() {
	stateLock.Lock()
	state.Upstream.Processing = true
	stateLock.Unlock()

	defer func() {
		stateLock.Lock()
		state.Upstream.Processing = false
		stateLock.Unlock()

		cancelMutex.Lock()
		cancelProcessing = nil
		cancelMutex.Unlock()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelMutex.Lock()
	cancelProcessing = cancel
	cancelMutex.Unlock()

	if err := process(ctx); err != nil {
		log.Printf("process quit: %s", err)
	}
}

func stopProcessing() {
	cancelMutex.Lock()
	defer cancelMutex.Unlock()

	if cancelProcessing != nil {
		cancelProcessing()
	}
}
