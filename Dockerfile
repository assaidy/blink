from golang:1.25-alpine as build
workdir /app
copy ./go.mod ./go.sum .
run go mod download
copy . .
run CGO_ENABLED=0 go build -ldflags="-s -w" -o api ./cmd/api/
run CGO_ENABLED=0 go build -ldflags="-s -w" -o workers ./cmd/workers/

from alpine:latest
workdir /app
copy --from=build /app/api /app/api
copy --from=build /app/workers /app/workers
cmd ["/app/api"]
