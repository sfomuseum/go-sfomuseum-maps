# go-sfomuseum-maps

Tools for working with maps defined in the sfomuseum-data/sfomuseum-data-maps repository.

## Important

Work in progress. Documentation is incomplete at this time.

### Years, dates and labels

Up to now most of the work involving historical aerial imagery of SFO has operated assuming a single map/image per year. Obviously, this isn't true but it made things simple at the time. This decision is reaching the limits of its usefulness and it is anticipated that there will be changes to output format of files like [dist/sfomuseum.maps.catalog.js](dist/sfomuseum.maps.catalog.js). Nothing has been decided yet but change(s) seem inevitable.

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/catalog_js cmd/catalog_js/main.go
go build -mod vendor -ldflags="-s -w" -o bin/qgis-tile-connections cmd/qgis-tile-connections/main.go
go build -mod vendor -ldflags="-s -w" -o bin/new cmd/new/main.go
```

### catalog_js

```
$> ./bin/catalog_js -h
Usage of ./bin/catalog_js:
  -exclude value
    	Zero or more maps to exclude (based on their sfomuseum:uri value)
  -mode string
    	A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator URI. (default "git://")
  -uri string
    	A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator source. (default "https://github.com/sfomuseum-data/sfomuseum-data-maps.git")
```

Generate a fresh `sfomuseum.maps.catalog.js` JavaScript package derived from sfomuseum-data-maps data and emit the output to `STDOUT`.

### qgis-tile-connections

```
$> ./bin/qgis-tile-connections -h
Usage of ./bin/qgis-tile-connections:
  -exclude value
    	Zero or more maps to exclude (based on their sfomuseum:uri value)
  -mode string
    	A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator URI. (default "git://")
  -uri string
    	A valid whosonfirst/go-whosonfirst-iterate-git/v2/iterator source. (default "https://github.com/sfomuseum-data/sfomuseum-data-maps.git")
```

Generate a fresh QGIS `qgsXYZTilesConnections` XML file from sfomuseum-data-maps data and emit the output to `STDOUT`.

## Distributions

### sfomuseum.maps.catalog.js

A JavaScript library using the package name `sfomuseum.maps.catalog` that defines SFO aerial map tiles and methods for working with.

### sfomuseum.maps.tileconnections.xml

A QGIS `qgsXYZTilesConnections` XML file that can be imported in to QGIS to load SFO aerial map "XYZ" layers.

_Important: The `qgsXYZTilesConnections` document definition does not provide any means to define extents for an XYZ layer so if you load any of these SFO aerial layers anywhere except the general area of SFO QGIS will start emitting a lot of tile requests that will fail._

## See also

* https://github.com/sfomuseum-data/sfomuseum-data-maps
* https://millsfield.sfomuseum.org/map
* https://github.com/sfomuseum/leaflet-layers-control
