from golang:1.25-alpine as build
workdir /app
copy ./go.mod ./go.sum .
run go mod download
copy . .
run CGO_ENABLED=0 go build -ldflags="-s -w" -o blink ./cmd/app/
run CGO_ENABLED=0 go build -ldflags="-s -w" -o workers ./cmd/workers/

from alpine:latest
workdir /app
copy --from=build /app/blink /app/blink
copy --from=build /app/workers /app/workers
cmd ["/app/blink"]
