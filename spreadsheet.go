package main

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	spreadsheet "gopkg.in/Iwark/spreadsheet.v2"
)

var (
	processSpreadsheetID = make(chan string, 16)
	currentDrivers       = []*Driver{}
	driversMap           = map[string]int{}
	driversLock          sync.RWMutex
)

type Driver struct {
	ID  int    `json:"id"`
	EPC string `json:"epc"`
}

func syncSpreadsheet() {
	service, err := spreadsheet.NewService()
	if err != nil {
		panic(err)
	}

	go func() {
		var (
			currentID          string
			currentSpreadsheet *spreadsheet.Spreadsheet
			currentSheet       *spreadsheet.Sheet
			sheetLock          sync.Mutex
		)

		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		for {
			select {
			case newID := <-processSpreadsheetID:
				log.Printf("spreadsheet: processing %s", newID)

				ss, err := service.FetchSpreadsheet(newID)
				if err != nil {
					log.Printf("spreadsheet: processing %s failed on fetch: %s", newID, err.Error())
					continue
				}
				sheet, err := ss.SheetByIndex(0)
				if err != nil {
					log.Printf("spreadsheet: processing %s failed on sheet: %s", newID, err.Error())
					continue
				}

				sheetLock.Lock()
				currentID = newID
				currentSpreadsheet = &ss
				currentSheet = sheet
				processSpreadsheet(currentSheet)
				sheetLock.Unlock()

			case <-ticker.C:
				if currentSpreadsheet == nil {
					continue
				}

				sheetLock.Lock()
				processSpreadsheetID <- currentID
				sheetLock.Unlock()
			}
		}
	}()
}

func processSpreadsheet(sheet *spreadsheet.Sheet) {
	var (
		idColumn  int = -1
		epcColumn int = -1
	)

	for i, field := range sheet.Rows[0] {
		if idColumn != -1 && epcColumn != -1 {
			break
		}

		if strings.Contains(field.Value, "numer") || field.Value == "id" {
			idColumn = i
			continue
		}

		if strings.Contains(field.Value, "epc") || strings.Contains(field.Value, "EPC") ||
			field.Value == "tag" {
			epcColumn = i
			continue
		}
	}

	newDrivers := []*Driver{}
	newMap := map[string]int{}
	for i, row := range sheet.Rows {
		if i == 0 || row[0].Value == "" {
			continue
		}

		parsedID, err := strconv.Atoi(row[idColumn].Value)
		if err != nil {
			continue
		}

		// Remove all spaces
		epcValue := strings.Replace(row[epcColumn].Value, " ", "", -1)
		if epcValue == "" {
			continue
		}

		newDrivers = append(newDrivers, &Driver{
			ID:  parsedID,
			EPC: epcValue,
		})

		newMap[epcValue] = parsedID
	}

	driversLock.Lock()
	stateLock.Lock()
	currentDrivers = newDrivers
	driversMap = newMap
	state.LegacySpreadsheet.Count = len(currentDrivers)
	stateLock.Unlock()
	driversLock.Unlock()

	log.Print("Reloaded the drivers list")
}
