# mod_tile rewritten in Golang

go_tile is a drop-in replacement for [mod_tile](https://github.com/openstreetmap/mod_tile).
It should work with both renderd and [tirex](https://github.com/openstreetmap/tirex), although development has thus far only been done with renderd.
go_tile is a static ~6MB binary with no external libraries, small memory footprint.

Currently supported features:
* serve prerendered tiles
* request tiles to be generated by renderd and then serve them

## Why was go_tile written?

* I want to make running a tileserver as easy as possible.
* mod_tile depends on Apache2 and requires an Apache2 config.
* mod_tile is not memory safe (cause it's written in C). This plants the lingering fear of buffer overflows and memory leaks
* Apache2+mod_tile use a thread per request, that should at some point introduce performance problems.
* mod_tile seems to be mostly in maintenance mode and not open for experimentation (there are many features in mod_tile that I don't plan to include in go_tile)

## Building from source

```shell
go build .
```

## Usage

### With a renderd backend

* Currently there are no binaries included here automatically (I might have manually included them for some releases), so you will need to build from source
* a slippymap with leaflet is provided in the ./static/ folder
* You need to have a working renderd setup

### With prerendered Tiles

With Docker:

```shell
docker run --rm -it -v $YOUR_TILE_FOLDER:/data -p 8080:8080 ghcr.io/nielsole/go_tile:latest
```

Now you can view your map at <http://localhost:8080/>. Tiles are served at <http://localhost:8080/>

### Usage options

If you prefer to run the binary directly you have the following options:

```
Usage of ./osm-tileserver:
  -data string
        Path to directory containing tiles (default "./data")
  -map string
        Name of map. This value is also used to determine the metatile subdirectory (default "ajt")
  -port string
        HTTP Listening port (default ":8080")
  -renderd-timeout int
        time in seconds to wait for renderd before returning an error to the client. Set negative to disable (default 60)
  -socket string
        Unix domain socket path or hostname:port for contacting renderd. Set to '' to disable rendering (default "/var/run/renderd/renderd.sock")
  -static string
        Path to static file directory (default "./static/")
  -tile_expiration duration
        Duration(example for a week: '168h') after which tiles are considered stale. Disabled by default
  -tls_cert_path string
        Path to TLS certificate
  -tls_key_path string
        Path to TLS key
  -tls_port string
        HTTPS Listening port. This listener is only enabled if both tls cert and key are set. (default ":8443")
```

## Pregenerate static tiles using mod_tile

1. Download data that you want <https://download.geofabrik.de/>
1. Follow the guides at <https://github.com/Overv/openstreetmap-tile-server> to get a working tile generation setup.
    * Make sure to persist the rendered tiles somewhere e.g. by mounting in `-v $YOUR_TILE_FOLDER:/var/lib/mod_tile`
2. Follow the [breadcrumbs](https://github.com/Overv/openstreetmap-tile-server/issues/15) about how to pregenerate tiles for the desired area and zoom-level.
    1. Get a shell in the container: `docker exec -it $CONTAINER_NAME bash`
    2. Download a script that makes it easier to specify GPS coordinates: `curl -o render_list_geo.pl https://raw.githubusercontent.com/alx77/render_list_geo.pl/master/render_list_geo.pl`
    3. Pregenerate the tiles to your liking: `perl ./render_list_geo.pl -m ajt -n 3 -z 6 -Z 16 -x 2.5 -X 6.5 -y 49.4 -Y 51.6`
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
