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
<head><title>Analytics Dashboard</title></head>
<body>
    <h1>Top Pages</h1>
    <ul>
        {{range .TopPages}}
        <li>{{.Path}}: <strong>{{.Views}} views</strong></li>
        {{end}}
    </ul>
</body>
</html>
`

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Query the DB for Top Pages
	topPagesQuery := `
			SELECT path, count(path) as views from hit
			where timestamp > ?
			group by path
			order by views desc
			limit 20;
		`
	topPages, err := DB.Exec(topPagesQuery, time.Now().Unix()-86400)
	if err != nil {
		log.Fatalf("failed to query db for top pages : %v", err)
	}
	// 2. Query the DB for Total Count
	totalCountQuery := `
			SELECT count(*) as total total_count from hit
			where timestamp > ?
	`
	totalCount := DB.QueryRow(totalCountQuery, time.Now().Unix()-86400)

	// 3. Render the template:
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	// data := ... (your queried data)
	// tmpl.Execute(w, data)
}
