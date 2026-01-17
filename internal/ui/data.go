package ui

import (
	"log"
	"time"

	"github.com/vimaurya/lumen/internal/storage"
)

func topPages() []PageStat {
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

	pages := []PageStat{}
	for topPages.Next() {
		var ps PageStat
		err := topPages.Scan(&ps.Path, &ps.Views)
		if err != nil {
			log.Fatal(err)
			continue
		}

		pages = append(pages, ps)
	}

	return pages
}

func totalCount() int64 {
	totalCountQuery := `
		SELECT count(*) as total_count from hit
		where timestamp > ?
	`
	var count int64
	err := storage.DB.QueryRow(totalCountQuery, time.Now().Unix()-86400).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	return count
}

func avgLatency() float32 {
	avgLatencyQuery := `
		SELECT COALESCE(AVG(duration), 0) FROM hit where timestamp > ?;
	`
	var avg float32

	err := storage.DB.QueryRow(avgLatencyQuery, time.Now().Unix()-86400).Scan(&avg)
	if err != nil {
		log.Fatal(err)
	}

	return avg
}

func errorRate() int {
	errorRateQuery := `
		SELECT COUNT(*) from hit where status >= 400 and timestamp > ?
	`

	var rate int
	err := storage.DB.QueryRow(errorRateQuery, time.Now().Unix()-86400).Scan(&rate)
	if err != nil {
		log.Fatal(err)
	}

	return rate
}

func uniqueVisitors() int64 {
	uniqueVisitorsQuery := `
		Select count(distinct hashuserid) from hit where timestamp > ?
	`
	var uniquevis int64

	err := storage.DB.QueryRow(uniqueVisitorsQuery, time.Now().Unix()-86400).Scan(&uniquevis)
	if err != nil {
		log.Fatal(err)
	}

	return uniquevis
}

func performance() []PerformanceStat {
	var performanceAnalysis []PerformanceStat

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

		performanceAnalysis = append(performanceAnalysis, ps)
	}

	return performanceAnalysis
}
