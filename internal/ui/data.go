package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/vimaurya/lumen/internal/storage"
)

func testData() {
	query := `INSERT INTO hit (path, sessionid, country, timestamp, isbot) 
				VALUES ('/test', 'session-123', 'India', 1737210000, 0),
				('/test', 'session-456', 'United States', 1737210000, 0),
				('/test', 'session-789', 'India', 1737210500, 0);`
	_, err := storage.DB.Exec(query)
	if err != nil {
		log.Printf("failed to input test data : %v", err)
	}
}

func fetchGlobalMetrics(data *DashboardData) {
	if storage.DB == nil {
		return
	}

	dayAgo := time.Now().Unix() - 86400
	fiveMinAgo := time.Now().Unix() - 300

	query := `
        SELECT 
            (SELECT COUNT(*) FROM hit WHERE timestamp > ?) as views,
            (SELECT COUNT(DISTINCT hashuserid) FROM hit WHERE timestamp > ?) as visitors,
            (SELECT COUNT(*) FROM hit WHERE status >= 400 AND timestamp > ?) as errors,
            (SELECT COUNT(DISTINCT sessionid) FROM hit WHERE timestamp > ? AND isbot = 0) as live
    `

	err := storage.DB.QueryRow(query, dayAgo, dayAgo, dayAgo, fiveMinAgo).Scan(
		&data.TotalCount,
		&data.UniqueVisitors,
		&data.ErrorRate,
		&data.Active,
	)
	if err != nil {
		log.Printf("Global metrics error: %v", err)
	}
}

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
	if storage.DB == nil {
		return
	}

	countQuery := `SELECT COUNT(DISTINCT sessionid) FROM hit WHERE isbot = 0`
	err := storage.DB.QueryRow(countQuery).Scan(uniqueSessions)
	if err != nil {
		log.Printf("Session Count Error: %v", err)
		return
	}

	timeQuery := `
		SELECT 
			COALESCE(
				(SUM(max_t) - SUM(min_t)) / CAST(COUNT(*) AS FLOAT), 
				0
			)
		FROM (
			SELECT MAX(timestamp) as max_t, MIN(timestamp) as min_t 
			FROM hit 
			WHERE isbot = 0 
			GROUP BY sessionid
		)`

	var avgSeconds float64
	err = storage.DB.QueryRow(timeQuery).Scan(&avgSeconds)
	if err != nil {
		log.Printf("Session Time Error: %v", err)
		*avgSessionTime = "0s"
	} else {
		*avgSessionTime = fmt.Sprintf("%.0fs", avgSeconds)
	}
}

func bounceRate(bounceRate *float64) {
	query := `
		SELECT 
			COALESCE(
				(CAST(COUNT(CASE WHEN hit_count = 1 THEN 1 END) AS FLOAT) / COUNT(*)) * 100, 
				0
			)
		FROM (
			SELECT COUNT(*) as hit_count 
			FROM hit 
			WHERE isbot = 0 
			GROUP BY sessionid
		)`

	err := storage.DB.QueryRow(query).Scan(bounceRate)
	if err != nil {
		log.Printf("bounceRate query err: %v", err)
		*bounceRate = 0.0
	}
}

func activeUsers(active *int) {
	if storage.DB == nil {
		return
	}

	fiveMinutesAgo := time.Now().Unix() - 300
	query := "SELECT COUNT(DISTINCT sessionid) FROM hit WHERE timestamp > ? AND isbot = 0"

	err := storage.DB.QueryRow(query, fiveMinutesAgo).Scan(active)
	if err != nil {
		log.Printf("activeUsers err: %v", err)
	}
}

func topCountries(stats *[]CountryStat) {
	if storage.DB == nil {
		return
	}

	query := `
		SELECT country, COUNT(*) as count 
		FROM hit 
		WHERE country != '' AND country IS NOT NULL
		GROUP BY country 
		ORDER BY count DESC 
		LIMIT 10`

	rows, err := storage.DB.Query(query)
	if err != nil {
		log.Printf("topCountries err: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cs CountryStat
		if err := rows.Scan(&cs.Country, &cs.Count); err == nil {
			*stats = append(*stats, cs)
		}
	}
}
