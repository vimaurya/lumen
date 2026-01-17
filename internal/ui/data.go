package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/vimaurya/lumen/internal/storage"
)

func topPages(pages *[]PageStat) {
	topPagesQuery := `
		SELECT path, count(path) as views from hit
		where timestamp > ?
		group by path
		order by views desc
		limit 20;
	`
	topPages, err := storage.DB.Query(topPagesQuery, time.Now().Unix()-86400)
	if err != nil {
		log.Fatalf("failed to query db for top pages : %v", err)
	}
	defer topPages.Close()

	for topPages.Next() {
		var ps PageStat
		err := topPages.Scan(&ps.Path, &ps.Views)
		if err != nil {
			log.Fatal(err)
			continue
		}

		*pages = append(*pages, ps)
	}
}

func totalCount(count *int64) {
	totalCountQuery := `
		SELECT count(*) as total_count from hit
		where timestamp > ?
	`
	err := storage.DB.QueryRow(totalCountQuery, time.Now().Unix()-86400).Scan(count)
	if err != nil {
		log.Fatal(err)
	}
}

func avgLatency(avg *float32) {
	avgLatencyQuery := `
		SELECT COALESCE(ROUND(CAST(AVG(duration) AS NUMERIC), 2), 0) 
		FROM hit 
		WHERE timestamp > ?;
	`
	err := storage.DB.QueryRow(avgLatencyQuery, time.Now().Unix()-86400).Scan(avg)
	if err != nil {
		log.Printf("avgLatency err : %v", err)
	}
}

func errorRate(rate *int) {
	errorRateQuery := `
		SELECT COUNT(*) from hit where status >= 400 and timestamp > ?
	`

	err := storage.DB.QueryRow(errorRateQuery, time.Now().Unix()-86400).Scan(rate)
	if err != nil {
		log.Fatal(err)
	}
}

func uniqueVisitors(uniquevis *int64) {
	uniqueVisitorsQuery := `
		Select count(distinct hashuserid) from hit where timestamp > ?
	`

	err := storage.DB.QueryRow(uniqueVisitorsQuery, time.Now().Unix()-86400).Scan(uniquevis)
	if err != nil {
		log.Fatal(err)
	}
}

func performance(performanceAnalysis *[]PerformanceStat) {
	performaceQuery := `
		SELECT path, COALESCE(AVG(duration), 0) as Avg_dur
		FROM hit
		WHERE timestamp > ?
		GROUP BY path 
		ORDER BY avg_dur DESC
		LIMIT 5
`
	rows, err := storage.DB.Query(performaceQuery, time.Now().Unix()-86400)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var ps PerformanceStat
		err = rows.Scan(&ps.Path, &ps.Avg_dur)
		if err != nil {
			log.Fatalf("scan error : %v", err)
			continue
		}

		*performanceAnalysis = append(*performanceAnalysis, ps)
	}
}

func uniqueSessions(uniqueSessions *int, avgSessionTime *string) {
	query := `
		Select	
			COUNT(DISTINCT sessionid),
			AVG(session_len)
			FROM (
				SELECT (MAX(timestamp) - MIN(timestamp)) as session_len 
				FROM hit 
				WHERE isbot = 0 
				GROUP BY sessionid
			)	
	`
	var avgSeconds float64
	err := storage.DB.QueryRow(query).Scan(uniqueSessions, &avgSeconds)
	if err != nil {
		log.Printf("uniqueSessions err : %v", err)
	}

	*avgSessionTime = fmt.Sprintf("%.0fs", avgSeconds)
}
