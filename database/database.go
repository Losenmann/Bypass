package database

import (
	"bypass/setup"
	"context"
	"database/sql"
	"errors"
	"os"
	//_ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

const (
	constCopm = 1
)

var (
	db     *sql.DB
	dbPath = *setup.DatabasePath
)

type Site struct {
	address string
	enable bool
}


func Check() {
	if _, err := os.Stat(dbPath); err == nil {
		setup.LoggingHendler(1, "Database is exist", constCopm)
	} else if errors.Is(err, os.ErrNotExist) {
		Init(dbPath)
	} else {
		setup.LoggingHendler(3, "Database other error", constCopm)
		os.Exit(constCopm)
	}
}

func Init(dbPath string) error {	
	if db, err := sql.Open("sqlite3", dbPath); err != nil {
		setup.LoggingHendler(3, "Database filed open: " + err.Error(), constCopm)
		os.Exit(constCopm)
	} else {
		defer db.Close()
		if resp, err := db.ExecContext(context.Background(), DBTemplateCreate); err != nil {
			setup.LoggingHendler(3, "Database error create tamplate: " + err.Error(), constCopm)
			os.Exit(constCopm)
		} else {
			setup.LoggingHendler(1, "Database created", constCopm)
			rowsa, _ := resp.RowsAffected()
			setup.LoggingHendler(1, rowsa, constCopm)
			return nil
		}
	}
	return nil
}

func Select() ([]Site) {
	if db, err := sql.Open("sqlite3", dbPath); err != nil {
        setup.LoggingHendler(2, "Database error select data: " + err.Error(), constCopm)
    } else {
		defer db.Close()
		if rows, err := db.QueryContext(context.Background(), QuerySelectAll); err != nil {
			setup.LoggingHendler(2, "Database error select data: " + err.Error(), constCopm)
		} else {
			defer rows.Close()
			sites := []Site{}
			for rows.Next() {
				site := Site{}
				if err := rows.Scan(&site.address, &site.enable); err != nil {
					setup.LoggingHendler(2, "Database error read data: " + err.Error(), constCopm)
					continue
				}
				sites = append(sites, site)
			}
			return sites
		}
	}
	return []Site{}
}