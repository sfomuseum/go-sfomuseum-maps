GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

REFRESH_MODE=git://
REFRESH_URI=https://github.com/sfomuseum-data/sfomuseum-data-maps.git

refresh:
	@make cli
	@make catalog.js
	@make tile_connections

refresh-local:
	@make cli
	@make catalog.js REFRESH_MODE=repo:// REFRESH_URI=/usr/local/data/sfomuseum-data-maps
	@make tile_connections  REFRESH_MODE=repo:// REFRESH_URI=/usr/local/data/sfomuseum-data-maps

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/catalog_js cmd/catalog_js/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/qgis-tile-connections cmd/qgis-tile-connections/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/new cmd/new/main.go

catalog.js:	
	./bin/catalog_js -mode $(REFRESH_MODE) -uri $(REFRESH_URI) > dist/sfomuseum.maps.catalog.js

tile_connections:
	./bin/qgis-tile-connections -mode $(REFRESH_MODE) -uri $(REFRESH_URI) > dist/sfomuseum.maps.tileconnections.xml
