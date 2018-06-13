package main

import "log"

func loadDefaultParams() {
	var autostartValue string
	if err := fdb.Read("params", "autostart", &autostartValue); err != nil {
		log.Printf("params: unable to read autostart: %s", err)
	}

	var tagValue string
	if err := fdb.Read("params", "tag", &tagValue); err != nil {
		log.Printf("params: unable to read tag: %s", err)
	}

	var convertValue string
	if err := fdb.Read("params", "convert", &convertValue); err != nil {
		log.Printf("params: unable to read convert: %s", err)
	}

	var validateValue string
	if err := fdb.Read("params", "validate", &validateValue); err != nil {
		log.Printf("params: unable to read the validate value: %s", err)
	}

	var legacySpreadsheetValue string
	if err := fdb.Read("params", "legacy_spreadsheet", &legacySpreadsheetValue); err != nil {
		log.Printf("params: unable to read the legacySpreadsheet value: %s", err)
	}

	var legacyAddressValue string
	if err := fdb.Read("params", "legacy_address", &legacyAddressValue); err != nil {
		log.Printf("params: unable to read the legacyAddress value: %s", err)
	}

	var upstreamValue string
	if err := fdb.Read("params", "upstream", &upstreamValue); err != nil {
		log.Printf("params: unable to read the upstream value: %s", err)
	}

	stateLock.Lock()
	state.Params.Autostart = autostartValue
	state.Params.Tag = tagValue
	state.Params.Convert = convertValue
	state.Params.Validate = validateValue
	state.Params.LegacySpreadsheet = legacySpreadsheetValue
	state.Params.LegacyAddress = legacyAddressValue
	state.Params.Upstream = upstreamValue
	stateLock.Unlock()

	go func() {
		processSpreadsheetID <- legacySpreadsheetValue
	}()
}

func setParam(name string, value string) {
	if name == "autostart" {
		if err := fdb.Write("params", "autostart", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.Autostart = value
		stateLock.Unlock()
		return
	}

	if name == "tag" {
		if err := fdb.Write("params", "tag", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.Tag = value
		stateLock.Unlock()
		return
	}

	if name == "convert" {
		if err := fdb.Write("params", "convert", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.Convert = value
		stateLock.Unlock()
		return
	}

	if name == "validate" {
		if err := fdb.Write("params", "validate", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.Validate = value
		stateLock.Unlock()
		return
	}

	if name == "legacy_spreadsheet" {
		if err := fdb.Write("params", "legacy_spreadsheet", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.LegacySpreadsheet = value
		stateLock.Unlock()

		go func() {
			processSpreadsheetID <- value
		}()
		return
	}

	if name == "legacy_address" {
		if err := fdb.Write("params", "legacy_address", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.LegacyAddress = value
		stateLock.Unlock()
		return
	}

	if name == "upstream" {
		if err := fdb.Write("params", "upstream", value); err != nil {
			panic(err)
		}

		stateLock.Lock()
		state.Params.Upstream = value
		stateLock.Unlock()
		return
	}
}
