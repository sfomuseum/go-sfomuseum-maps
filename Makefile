cli:
	go build -mod vendor -o bin/catalog_js cmd/catalog_js/main.go
	go build -mod vendor -o bin/new cmd/new/main.go

catalog.js:	
	./bin/catalog_js > dist/sfomuseum.maps.catalog.js
