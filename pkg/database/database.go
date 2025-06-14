package database

import (
	"fmt"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

type DbID uint

const (
	SIG_DB DbID = iota
	RESULTS_DB
)

type Configuration struct {
	fpath  string
	config *gorm.Config
	models []any
}

var dbs = [...]Configuration{
	{
		fpath: "signatures.db",
		config: &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		},
		models: []any{&Signature{}, &Rule{}, &Scan{}, &Probe{}},
	},
	{
		// TODO: Add prometheus plugin to this one
		// db.Use(prometheus.New(prometheus.Config{}))
		fpath: "results.db",
		config: &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		},
		models: []any{&Host{}, &Record{}, &Label{}, &Mark{}},
	},
}

func GetDatabase(id DbID) Configuration {
	return dbs[id]
}

func Open(conf Configuration) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(conf.fpath), conf.config)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s", conf.fpath)
	}

	db = db.Exec("PRAGMA foreign_keys = ON")
	if err := db.AutoMigrate(conf.models...); err != nil {
		return nil, err
	}

	return db, nil
}
