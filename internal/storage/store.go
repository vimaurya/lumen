package storage

import (
	"database/sql"
	"log"
	"time"

	"github.com/vimaurya/lumen/internal/analytics"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite", "analytics.db")
	if err != nil {
		return err
	}
	DB.Exec("journal_mode=WAL;")
	DB.Exec("synchronous=NORMAL;")

	query := `
		create table if not exists hit (
			path varchar,
			hashuserid varchar,
			referrer varchar,
			timestamp INTEGER,
			country varchar,
			browser varchar,
			device varchar,
			duration INTEGER,
			operating_system varchar,
			status INTEGER,
			method varchar,
			requestsize INTEGER,
			sessionid TEXT,
			isbot boolean
		)
	`
	_, err = DB.Exec(query)
	if err != nil {
		log.Printf("failed to create table : %v", err)
	}
	return err
}

func StartWorker() {
	batch := make([]analytics.Hit, 0, 100)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case hit, ok := <-analytics.HitBuffer:
			if !ok && len(batch) > 0 {
				flush(batch)
				return
			}
			batch = append(batch, hit)
			if len(batch) > 100 {
				err := flush(batch)
				if err != nil {
					log.Printf("failed to flush into database... %v", err)
				}
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				err := flush(batch)
				if err != nil {
					log.Printf("failed to flush into database... %v", err)
				}
				batch = batch[:0]
			}
		}
	}
}

func flush(batch []analytics.Hit) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	query := `INSERT INTO hit(
				path, hashuserid, referrer, timestamp, country, 
			browser, device, duration, operating_system, status,
			method, requestsize, sessionid, isbot
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, hit := range batch {
		_, err := tx.Exec(query,
			hit.Path,
			hit.HashedUserId,
			hit.Referrer,
			hit.Timestamp,
			hit.Country,
			hit.Browser,
			hit.Device,
			hit.Duration,
			hit.OperatingSystem,
			hit.Status,
			hit.Method,
			hit.RequestSize,
			hit.SessionId,
			hit.IsBot,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
