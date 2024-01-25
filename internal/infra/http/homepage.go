package http

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type homeHandler struct {
	content []byte
}

type homeInformation struct {
	Version    string
	MetricPath string
}

func NewHomePage(version string, metricPath string) (*homeHandler, error) {
	context := homeInformation{
		Version:    version,
		MetricPath: metricPath,
	}

	content := `<html>
	<head>
		<title>Prometheus RDS Exporter</title>
	</head>
	<body>
		<h1>Prometheus RDS Exporter ({{ .Version }})</h1>
		<p><a href='{{ .MetricPath }}'>Metrics</a></p>
	</body>
</html>`

	homepage := homeHandler{}

	tmpl, err := template.New("homepage").Parse(content)
	if err != nil {
		return &homepage, fmt.Errorf("failed to load template: %w", err)
	}

	renderedHTMLBuffer := new(bytes.Buffer)

	err = tmpl.Execute(renderedHTMLBuffer, context)
	if err != nil {
		return &homepage, fmt.Errorf("failed to render homepage: %w", err)
	}

	homepage.content = renderedHTMLBuffer.Bytes()

	return &homepage, nil
}

func (h homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	_, _ = w.Write(h.content) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter // h.content is rendered by html/template in constructor
}
