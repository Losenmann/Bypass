package main

import (
	"fmt"
	"database/sql"
	_ "modernc.org/sqlite"
	"context"
)

var (
	db *sql.DB
	dbPath = "./databse.db"
)

func main() {
    fmt.Println("Hello, World!")
	initDatabase(dbPath)
}

func initDatabase(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS tb_category
		(
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		name TEXT DEFAULT (strftime('default-cat-%s', 'now')) UNIQUE,
    		description TEXT,
    		enable INTEGER DEFAULT 1 CHECK(enable =0 OR enable =1)
		);
		CREATE TABLE IF NOT EXISTS tb_resource
		(
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		name TEXT DEFAULT (strftime('default-res-%s', 'now')) UNIQUE,
    		description TEXT,
    		enable INTEGER DEFAULT 1 CHECK(enable =0 OR enable =1)
		);
		CREATE TABLE IF NOT EXISTS tb_site
		(
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		address TEXT NOT NULL,
    		cat INTEGER REFERENCES tb_category(id) DEFAULT 0,
    		res INTEGER REFERENCES tb_resource(id) DEFAULT 0,
    		enable INTEGER DEFAULT 1 CHECK(enable =0 OR enable =1),
    		timeadd INTEGER DEFAULT CURRENT_TIMESTAMP
		);
		DROP VIEW IF EXISTS vw_site;
		CREATE VIEW IF NOT EXISTS vw_site AS
    		SELECT ts.address,
        		tc.name As cat,
        		tr.name AS res,
        		ts.enable,
        		ts.timeadd
    		FROM tb_site ts
    		INNER JOIN tb_category tc ON ts.cat = tc.id
    		INNER JOIN tb_resource tr ON ts.res = tr.id;
		INSERT INTO tb_category(name,description) VALUES('default','Default category'),('Social','Social networks'),('Media','Videos or music'),('Entertainment','Movies, books'),('Game','Game servers, game sites'),('Dev','Repositories, image registry, documentation'),('File Exchangers','File sharing, cloud drives');
		INSERT INTO tb_resource(name,description) VALUES('default','Default resource');
		`,
	)
	if err != nil {
		return err
	}
	return nil
}