# Dead simple OSM static tileserver

go_tile serves prerendered OpenStreetMap tiles from the local filesystem prerendered by mod_tile.
go_tile is a static binary based on 100 lines of Go which makes it **easier to setup / reason about** than Apache+mod_tile+renderd+mapnik. You don't have to worry about tile expirations, render configurations, database connections... at runtime(but of course during prerendering of the tiles).

The cache files created by mod_tile are `METATILE`s which can't be served directly through a webserver or e.g. S3.
Each tile in OSM is 64x64 pixels with many tiles only having ~100 bytes. mod_tile uses `METATILE`s to reduce the number of small files by grouping 8x8 tiles into one file.

## Usage

With Docker:

```
docker run --rm -it -v $YOUR_TILE_FOLDER:/data -p 8080:8080 nielsole/go_tile
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