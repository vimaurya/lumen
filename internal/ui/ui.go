package ui

import (
	"html/template"
	"net/http"
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
	<div class="main-card">
			<h3>System Health (Infrastructure)</h3>
			<div class="grid">
					<div class="stat-card"><h3>Views</h3><p>{{.TotalCount}}</p></div>
					<div class="stat-card"><h3>Avg Speed</h3><p>{{printf "%.2f" .AvgLatency}}ms</p></div>
					<div class="stat-card"><h3>Uniques (IP)</h3><p>{{.UniqueVisitors}}</p></div>
					<div class="stat-card"><h3 class="error-text">Errors</h3><p class="error-text">{{.ErrorRate}}</p></div>
			</div>
	</div>

	<div class="main-card">
			<h3>Visitor Engagement (Behavior)</h3>
			<div class="grid">
					<div class="stat-card"><h3>Visits (Sessions)</h3><p>{{.UniqueSessions}}</p></div>
					<div class="stat-card"><h3>Avg Session Time</h3><p>{{.AvgSessionTime}}</p></div>
					<div class="stat-card"><h3>Bounce Rate</h3><p>{{printf "%.1f" .BounceRate}}%</p></div>
					<div class="stat-card"><h3>Active Users</h3><p>{{.Active}}</p></div> </div>
	</div>

	<div class="main-card">
		<h2>Top Content (Performance & Popularity)</h2>
			<table>
					<thead>
							<tr>
									<th>URL Path</th>
									<th>Views</th>
							</tr>
					</thead>
					<tbody>
							{{range .TopPages}}
							<tr>
									<td>{{.Path}}</td>
									<td><strong>{{.Views}}</strong></td>
							</tr>
							{{end}}
					</tbody>
			</table>
	</div>

	<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>

	<div class="main-card">
			<h3>Top 10 Countries</h3>
			<div style="max-height: 300px;">
					<canvas id="countryChart"></canvas>
			</div>
	</div>

	<script>
			const ctx = document.getElementById('countryChart').getContext('2d');
			new Chart(ctx, {
					type: 'bar',
					data: {
							labels: [{{range .TopCountries}} "{{.Country}}", {{end}}],
							datasets: [{
									label: 'Visitors',
									data: [{{range .TopCountries}} {{.Count}}, {{end}}],
									backgroundColor: '#007bff'
							}]
					},
					options: {
							indexAxis: 'y',
							responsive: true,
							maintainAspectRatio: false
					}
			});
	</script>

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

type CountryStat struct {
	Country string
	Count   int
}
type DashboardData struct {
	TopPages       []PageStat
	Performance    []PerformanceStat
	TopCountries   []CountryStat
	TotalCount     int64
	UniqueVisitors int64
	AvgLatency     float32
	ErrorRate      int
	CurrentTime    string
	UniqueSessions int
	AvgSessionTime string
	BounceRate     float64
	Active         int
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := DashboardData{}

	testData()

	topPages(&data.TopPages)

	fetchGlobalMetrics(&data)

	performance(&data.Performance)

	topCountries(&data.TopCountries)

	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	tmpl.Execute(w, data)
}
