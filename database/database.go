package database

import (
	"speedtest/config"
	"speedtest/database/bolt"
	"speedtest/database/schema"
)

var (
	DB DataAccess
)

type DataAccess interface {
	SaveTelemetry(*schema.TelemetryData) error
	GetTelemetryByUUID(string) (*schema.TelemetryData, error)
	GetLastNRecords(int) ([]schema.TelemetryData, error)
	GetAllTelemetry() ([]schema.TelemetryData, error)
}

func SetDBInfo(cfg *config.Config) {
	DB = bolt.OpenDatabase(cfg.Database.Path)
}
