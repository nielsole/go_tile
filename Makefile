

build: osm-tileserver
	go build .

# docker run -p 8080:8080 -p 8081:80 -v osm-data:/data/database/ --name osm-test -d overv/openstreetmap-tile-server run
test: build
	docker cp ./osm-tileserver osm-test:/bin/osm-tileserver
	docker cp ./static/ osm-test:/static
	docker exec -it osm-test /bin/osm-tileserver -static /static -data /var/lib/mod_tile/ajt
