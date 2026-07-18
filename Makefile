GOOS=linux
GOARCH=amd64
LAMBDA_BIN=dist/%/bootstrap

.PHONY: build clean deploy

build: dist/resource-alert/bootstrap dist/weekly-bill/bootstrap dist/mid-month-forecast/bootstrap

dist/resource-alert/bootstrap:
	mkdir -p dist/resource-alert
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ ./cmd/resource-alert

dist/weekly-bill/bootstrap:
	mkdir -p dist/weekly-bill
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ ./cmd/weekly-bill

dist/mid-month-forecast/bootstrap:
	mkdir -p dist/mid-month-forecast
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ ./cmd/mid-month-forecast

deploy: build
	cd terraform && terraform init && terraform apply

clean:
	rm -rf dist
