# Build binary
FROM golang:alpine as build

WORKDIR /build

COPY go.mod go.sum ./
COPY . .


RUN GOOS=linux go build -ldflags="-w -s" -o ./main main.go

FROM scratch


WORKDIR "/app"
COPY --from=build /build/main /app
COPY .env /app/.env


CMD ["/app/main"]