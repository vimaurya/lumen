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
        body { font-family: sans-serif; margin: 40px; background: #fafafa; }
        .card { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { text-align: left; padding: 12px; border-bottom: 1px solid #ddd; }
        th { background: #eee; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Site Stats</h1>
        <p><strong>Total Page Views (24h):</strong> {{.TotalCount}}</p>
        
        <h2>Top Pages</h2>
        <table>
            <thead>
                <tr><th>URL Path</th><th>Views</th></tr>
            </thead>
            <tbody>
                {{range .TopPages}}
                <tr>
                    <td>{{.Path}}</td>
                    <td>{{.Views}}</td>
                </tr>
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

type DashboardData struct {
	TopPages    []PageStat
	TotalCount  int
	CurrentTime string
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
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

	totalCountQuery := `
		SELECT count(*) as total_count from hit
		where timestamp > ?
	`

	var totalCount int
	err = DB.QueryRow(totalCountQuery, time.Now().Unix()-86400).Scan(&totalCount)
	if err != nil {
		log.Fatal(err)
	}

	data := DashboardData{
		TopPages:    pages,
		TotalCount:  totalCount,
		CurrentTime: time.Now().Format("2006-01-02"),
	}

	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	tmpl.Execute(w, data)
}
