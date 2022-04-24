FROM golang:1.17-alpine as build
WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o server .
FROM scratch
COPY --from=build /build/server .
ADD ./static ./static
ENTRYPOINT ["./server"]
