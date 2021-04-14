FROM golang:1.16-alpine as build
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o server .
FROM scratch
COPY --from=build /build/server .
ENTRYPOINT ["./server"]
