package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	db, err := sql.Open("sqlite3", "analytics.db")
	if err != nil {
		return err
	}

	db.Exec("journal_mode=WAL;")
	db.Exec("synchronous=NORMAL;")

	query := `
		create table hit (
			path varchar,
			hashuserid varchar,
			referrer varchar,
			timestamp INTEGER 
		)
	`
	_, err = db.Exec(query)

	DB = db
	return err
}

func StartWorker() {
	batch := make([]Hit, 0, 100)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case hit, ok := <-HitBuffer:
			if !ok && len(batch) > 0 {
				flush(batch)
				return
			}
			batch = append(batch, hit)
			if len(batch) > 100 {
				err := flush(batch)
				if err != nil {
					log.Fatal("failed to flush into database...")
				}
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				err := flush(batch)
				if err != nil {
					log.Fatal("failed to flush into database...")
				}
				batch = batch[:0]
			}
		}
	}
}

func flush(batch []Hit) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	for _, hit := range batch {
		query := `
			INSERT INTO hit(path, hashuserid, referrer, timestamp) VALUES(?, ?, ?, ?)
		`

		_, err := tx.Exec(query, hit.Path, hit.HashedUserId, hit.Referrer, hit.Timestamp)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
