package main

import (
	"encoding/json"

	"github.com/lum8rjack/redcompass/services/namecheap"
	"github.com/pocketbase/pocketbase/core"
)

func SetupHooks() {
	bootstrapHook()
	namecheapCreateHook()
	namecheapUpdateHook()
	namecheapDeleteHook()
}

func bootstrapHook() {
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		namecheapStartupHook()

		return nil
	})
}

func namecheapStartupHook() {
	msg := "CRON:Namecheap startup hook"

	record, err := app.FindFirstRecordByData("Services", "Provider", "Namecheap")
	if err != nil {
		return
	}

	recordSettings := record.GetString("Settings")
	if recordSettings == "" {
		app.Logger().Error(msg, "function", "record.GetString", "error", "no settings")
		return
	}

	var settings namecheap.Settings
	err = json.Unmarshal([]byte(recordSettings), &settings)
	if err != nil {
		app.Logger().Error(msg, "function", "json.Unmarshal", "error", err.Error())
		return
	}

	if settings.ApiKey == "" || settings.IP == "" || settings.Username == "" {
		emtpyKey := settings.ApiKey == ""
		emtpyIP := settings.IP == ""
		emtpyUsername := settings.Username == ""

		app.Logger().Error(msg, "error", "invalid settings", "emptyKey", emtpyKey, "emptyIP", emtpyIP, "emptyUsername", emtpyUsername)
		return
	}

	recordCron := record.GetString("Cron")
	if recordCron == "" {
		app.Logger().Error(msg, "function", "record.GetString", "error", "no cron")
		return
	}

	AddNamecheapCron("Namecheap", recordCron)
	app.Logger().Info(msg, "status", "created", "jobID", "Namecheap", "cron", recordCron)
}

func namecheapDeleteHook() {
	app.OnRecordDelete("Services").BindFunc(func(e *core.RecordEvent) error {
		msg := "CRON:Namecheap delete hook"

		recordName := e.Record.GetString("Provider")
		if recordName != "Namecheap" {
			return e.Next()
		}

		RemoveNamecheapCron("Namecheap")
		app.Logger().Info(msg, "status", "deleted", "jobID", "Namecheap")
		return e.Next()
	})
}

func namecheapCreateHook() {
	app.OnRecordAfterCreateSuccess("Services").BindFunc(func(e *core.RecordEvent) error {
		msg := "CRON:Namecheap create hook"

		namecheapHelperFunc(e, msg)

		return e.Next()
	})
}

func namecheapUpdateHook() {
	app.OnRecordAfterUpdateSuccess("Services").BindFunc(func(e *core.RecordEvent) error {
		msg := "CRON:Namecheap update hook"

		// Check if the record Provider was changed / no Namecheap providers
		_, err := app.FindFirstRecordByData("Services", "Provider", "Namecheap")
		if err != nil {
			app.Logger().Error(msg, "function", "app.FindFirstRecordByData", "error", err.Error()) // sql: no rows in result set
			RemoveNamecheapCron("Namecheap")
			return e.Next()
		}

		namecheapHelperFunc(e, msg)
		return e.Next()
	})
}

func namecheapHelperFunc(e *core.RecordEvent, msg string) {
	// Remove the old cron job
	RemoveNamecheapCron("Namecheap")
	app.Logger().Info(msg, "status", "deleted", "jobID", "Namecheap")

	// Check if the record Provider was changed / no Namecheap providers
	recordName := e.Record.GetString("Provider")
	if recordName != "Namecheap" {
		return
	}

	// Get the new settings
	recordSettings := e.Record.GetString("Settings")
	if recordSettings == "" {
		app.Logger().Error(msg, "function", "record.GetString", "error", "no settings")
		return
	}

	// Unmarshal the settings
	var settings namecheap.Settings
	err := json.Unmarshal([]byte(recordSettings), &settings)
	if err != nil {
		app.Logger().Error(msg, "function", "json.Unmarshal", "error", err.Error())
		return
	}

	// Check if the settings are valid
	if settings.ApiKey == "" || settings.IP == "" || settings.Username == "" {
		emptyKey := settings.ApiKey == ""
		emptyIP := settings.IP == ""
		emptyUsername := settings.Username == ""

		app.Logger().Error(msg, "error", "invalid settings", "emptyKey", emptyKey, "emptyIP", emptyIP, "emptyUsername", emptyUsername)
		return
	}

	// Get the new cron
	recordCron := e.Record.GetString("Cron")
	if recordCron == "" {
		app.Logger().Error(msg, "function", "record.GetString", "error", "no cron")
		return
	}

	// Add the new cron job
	AddNamecheapCron("Namecheap", recordCron)
	app.Logger().Info(msg, "status", "created", "jobID", "Namecheap", "cron", recordCron)
}
