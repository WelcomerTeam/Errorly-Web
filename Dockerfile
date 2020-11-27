FROM golang:1.15-alpine AS build_base

RUN apk add --no-cache git build-base

WORKDIR /tmp/errorly

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/errorly ./cmd/main.go

FROM alpine:3.9
RUN apk add ca-certificates

COPY --from=build_base /tmp/errorly/out/errorly /app/errorly
COPY --from=build_base /tmp/errorly/web/dist /web/dist

EXPOSE 8001
CMD ["/app/errorly"]
