kubeconform_command := kubeconform -kubernetes-version $${KUBERNETES_VERSION-1.25.0} -cache $${KUBECONFORM_CACHE_DIRECTORY-/tmp} -summary -exit-on-error --strict -schema-location default -schema-location 'kubeconform/{{ .ResourceKind }}{{ .KindSuffix }}.json' -schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'
BUILD_INFO_PACKAGE_PATH=github.com/qonto/prometheus-rds-exporter/internal/infra/build
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
RELEASE_VERSION=$(shell jq .tag dist/metadata.json)
GIT_COMMIT_SHA=$(shell git rev-parse HEAD)
AWS_ECR_PUBLIC_ORGANIZATION=g1r8z6f4
BINARY=prometheus-rds-exporter
HELM_CHART_NAME=prometheus-rds-exporter-chart

all: build

.PHONY: format
format:
	gofumpt -l -w .

.PHONY: build
build:
	CGO_ENABLED=0 go build -v -ldflags="-X '$(BUILD_INFO_PACKAGE_PATH).Version=development' -X '$(BUILD_INFO_PACKAGE_PATH).CommitSHA=$(GIT_COMMIT_SHA)' -X '$(BUILD_INFO_PACKAGE_PATH).Date=$(BUILD_DATE)'" -o $(BINARY)

.PHONY: run
run:
	./$(BINARY) $(args)

.PHONY: test
test:
	go test -race -v ./... -coverprofile=coverage.txt -covermode atomic
	go install github.com/boumenot/gocover-cobertura@latest
	go run github.com/boumenot/gocover-cobertura@latest < coverage.txt > coverage.xml

.PHONY: lint
lint:
	golangci-lint run --verbose --timeout 2m

.PHONY: helm-test
helm-test:
	helm unittest configs/helm

.PHONY: helm-release
helm-release:
	./scripts/helm-release.sh $(HELM_CHART_NAME) configs/helm $(RELEASE_VERSION) $(AWS_ECR_PUBLIC_ORGANIZATION)

.PHONY: kubeconform
kubeconform:
	helm template configs/helm | $(kubeconform_command)

.PHONY: metrics-list
metrics-list:
	echo "| Name | Description |" > metrics
	echo "| ------ | ----------- |" >> metrics
	curl -s localhost:9043/metrics | grep -E '^# HELP' | awk '{metric = $$3; $$1=$$2=$$3=""; print "| " metric " | " $$0 " | "}' | sed -e's/  */ /g' >> metrics

.PHONY: all-tests
all-tests: test kubeconform helm-test
