FROM golang:1.20-alpine as build
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o server .

FROM busybox
COPY --from=build /build/server /bin/go_tile
ADD ./static /usr/share/go_tile/static
ENTRYPOINT ["/bin/go_tile", "--static", "/usr/share/go_tile/static"]
