# Dead simple OSM static tileserver

go_tile serves OpenStreetMap tiles from the local filesystem prerendered by mod_tile.
go_tile is a static ~6MB binary with minimal memory footprint based on 100 lines of Go.

## Which problem does this solve?

When running Apache+mod_tile+renderd+mapnik you have to worry about tile expirations, render configurations, database connections which is not trivial.
Running a tile server is has hardware requirements that are prohibitive for small embedded projects(RPis) and VPSes.

The cache files created by mod_tile can't be served directly through a webserver or CDN.
mod_tile uses `METATILE`s to reduce the number of small files by grouping 8x8 tiles into one file.

## Usage

With Docker:

```
docker run --rm -it -v $YOUR_TILE_FOLDER:/data -p 8080:8080 ghcr.io/nielsole/go_tile:latest
```

Now you can view your map at <http://localhost:8080/>. Tiles are served at <http://localhost:8080/>

If you prefer to run the binary directly you have the following options:

```
Usage of /tmp/go-build004735986/b001/exe/osm-tileserver:
  -data string
        Path to directory containing tiles (default "./data")
  -port string
        Listening port (default ":8080")
  -static string
        Path to static file directory (default "./static/")
```

## Generate tiles using mod_tile (Prerequisite)

1. Download data that you want <https://download.geofabrik.de/>
1. Follow the guides at <https://github.com/Overv/openstreetmap-tile-server> to get a working tile generation setup.
    * Make sure to persist the rendered tiles somewhere e.g. by mounting in `-v $YOUR_TILE_FOLDER:/var/lib/mod_tile`
2. Follow the [breadcrumbs](https://github.com/Overv/openstreetmap-tile-server/issues/15) about how to pregenerate tiles for the desired area and zoom-level.
    1. Get a shell in the container: `docker exec -it $CONTAINER_NAME bash`
    2. Download a script that makes it easier to specify GPS coordinates: `curl -o render_list_geo.pl https://raw.githubusercontent.com/alx77/render_list_geo.pl/master/render_list_geo.pl`
    3. Pregenerate the tiles to your liking: `perl ./render_list_geo.pl -m ajt  -n 3 -z 6 -Z 16 -x 2.5 -X 6.5 -y 49.4 -Y 51.6`
3. Cheers, you now have tiles in the `$YOUR_TILE_FOLDER`.

## Performance

The below was run on a 10 year old `Intel(R) Core(TM) i5-3450 CPU @ 3.10GHz`:

```
$ ab -n 100000 -c 12 -k internalip:8080/tile/17/70414/42993.png
...
Document Path:          /tile/17/70414/42993.png
Document Length:        18463 bytes

Concurrency Level:      12
Time taken for tests:   3.660 seconds
Complete requests:      100000
Failed requests:        0
Keep-Alive requests:    100000
Total transferred:      1868400000 bytes
HTML transferred:       1846300000 bytes
Requests per second:    27321.78 [#/sec] (mean)
Time per request:       0.439 [ms] (mean)
Time per request:       0.037 [ms] (mean, across all concurrent requests)
Transfer rate:          498515.71 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       1
Processing:     0    0   0.4      0       9
Waiting:        0    0   0.4      0       8
Total:          0    0   0.4      0       9

Percentage of the requests served within a certain time (ms)
  50%      0
  66%      0
  75%      1
  80%      1
  90%      1
  95%      1
  98%      1
  99%      2
 100%      9 (longest request)
```

This benchmark doesn't access the disk, as the tile has obviously been cached in memory. 
Anyways it should give you an indication of whether this is fast enough for your use-case.
