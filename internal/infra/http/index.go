package http

import (
	"fmt"
	"net/http"

	"github.com/qonto/prometheus-rds-exporter/internal/infra/build"
)

type helloWorldhandler struct{}

func (h helloWorldhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html>
		<head>
			<title>Prometheus RDS Exporter</title>
		</head>
		<body>
			<h1>Prometheus RDS Exporter (%s)</h1>
			<p><a href='/metrics'>Metrics</a></p>
		</body>
		</html>`, build.Version)
}
