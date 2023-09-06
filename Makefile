GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/catalog_js cmd/catalog_js/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/new cmd/new/main.go

catalog.js:	
	./bin/catalog_js > dist/sfomuseum.maps.catalog.js
