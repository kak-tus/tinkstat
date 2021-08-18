package main

const templateData = `
<html>

<head>
	<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/luxon"></script>
  <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-luxon"></script>
	<style>
	body {
		display: grid;
		grid-template-rows: repeat(3, 1fr);
		grid-template-columns: repeat(5, 1fr);
		grid-gap: 2vw;
	}
	</style>
</head>

<body>
{{ range $i, $chart := .Charts }}
	<div>
		<canvas id="chart{{$i}}"></canvas>
	</div>
{{- end }}

	<script>
	{{range $i, $chart := .Charts}}
		var labels{{$i}} = [{{ range $chart.Labels }}{{.}},{{end}}];

		var data{{$i}} = {
			labels: labels{{$i}},
			datasets: [{
				label: '{{$chart.Ticker}} {{$chart.From}}',
				backgroundColor: 'rgb(0, 0, 0)',
				borderColor: 'rgb(0, 0, 0)',
				data: [{{ range $chart.Points }}{{.}},{{end}}],
			}]
		};

		var config{{$i}} = {
			type: 'line',
			data: data{{$i}},
			options: {
				animation: false,
				normalized: true,
				spanGaps: true,
				scales: {
					x: {
						type: 'time'
					}
				},
				elements: {
					point: {
						radius: 0
					}
				}
			}
		};

		var myChart{{$i}} = new Chart(
			document.getElementById('chart{{$i}}'),
			config{{$i}}
		);
		{{end}}
	</script>
</body>

</html>
`
