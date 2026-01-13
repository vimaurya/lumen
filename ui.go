package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

var dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Analytics Dashboard</title>
    <style>
        body { font-family: 'Inter', sans-serif; margin: 40px; background: #f0f2f5; color: #333; }
        .grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 30px; }
        .stat-card { background: white; padding: 20px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); text-align: center; }
        .stat-card h3 { margin: 0; font-size: 14px; color: #666; text-transform: uppercase; }
        .stat-card p { margin: 10px 0 0; font-size: 24px; font-weight: bold; color: #007bff; }
        .main-card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; margin-top: 15px; }
        th, td { text-align: left; padding: 12px; border-bottom: 1px solid #eee; }
        .error-text { color: #dc3545; }
    </style>
</head>
<body>
    <h1>System Insights</h1>
    
    <div class="grid">
        <div class="stat-card"><h3>Views</h3><p>{{.TotalCount}}</p></div>
        <div class="stat-card"><h3>Uniques</h3><p>{{.UniqueVisitors}}</p></div>
        <div class="stat-card"><h3>Avg Speed</h3><p>{{.AvgLatency}}ms</p></div>
        <div class="stat-card"><h3 class="error-text">Errors</h3><p class="error-text">{{.ErrorRate}}</p></div>
    </div>

    <div class="main-card">
        <h2>Slowest Endpoints (Performance Debt)</h2>
        <table>
            <thead><tr><th>Path</th><th>Avg Latency</th></tr></thead>
            <tbody>
                {{range .Performance}}
                <tr><td>{{.Path}}</td><td><strong>{{printf "%.2f" .Avg_dur}}ms</strong></td></tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <div class="main-card">
        <h2>Top Content</h2>
        <table>
            <thead><tr><th>URL Path</th><th>Views</th></tr></thead>
            <tbody>
                {{range .TopPages}}
                <tr><td>{{.Path}}</td><td>{{.Views}}</td></tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>
`

type PageStat struct {
	Path  string
	Views int
}

type PerformanceStat struct {
	Path    string
	Avg_dur float64
}

type DashboardData struct {
	TopPages       []PageStat
	Performance    []PerformanceStat
	TotalCount     int
	UniqueVisitors int
	AvgLatency     float64
	ErrorRate      int
	CurrentTime    string
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := DashboardData{}

	topPagesQuery := `
		SELECT path, count(path) as views from hit
		where timestamp > ?
		group by path
		order by views desc
		limit 20;
	`
	topPages, err := DB.Query(topPagesQuery, time.Now().Unix()-86400)
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
		}

		pages = append(pages, ps)
	}

	data.TopPages = pages

	totalCountQuery := `
		SELECT count(*) as total_count from hit
		where timestamp > ?
	`

	err = DB.QueryRow(totalCountQuery, time.Now().Unix()-86400).Scan(&data.TotalCount)
	if err != nil {
		log.Fatal(err)
	}

	avgLatencyQuery := `
		SELECT COALESCE(AVG(duration), 0) FROM hit;
	`

	err = DB.QueryRow(avgLatencyQuery).Scan(&data.AvgLatency)
	if err != nil {
		log.Fatal(err)
	}

	errorRateQuery := `
		SELECT COUNT(*) from hit where status >= 400 and timestamp > ?
	`
	err = DB.QueryRow(errorRateQuery, time.Now().Unix()-86400).Scan(&data.ErrorRate)
	if err != nil {
		log.Fatal(err)
	}

	uniqueVisitorsQuery := `
		Select count(distinct hashuserid) from hit where timestamp > ?
	`
	err = DB.QueryRow(uniqueVisitorsQuery, time.Now().Unix()-86400).Scan(&data.UniqueVisitors)
	if err != nil {
		log.Fatal(err)
	}
	var performanceAnalysis []PerformanceStat

	performaceQuery := `
		SELECT path, COALESCE(AVG(duration), 0) as Avg_dur
		FROM hit
		WHERE timestamp > ?
		GROUP BY path 
		ORDER BY avg_dur DESC
		LIMIT 5
`
	rows, err := DB.Query(performaceQuery, time.Now().Unix()-86400)
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

	data.Performance = performanceAnalysis

	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	tmpl.Execute(w, data)
}
