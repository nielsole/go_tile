FROM golang:1.20-alpine as build
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o server .
FROM busybox
COPY --from=build /build/server .
ADD ./static ./static
ENTRYPOINT ["./server"]
