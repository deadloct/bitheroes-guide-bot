FROM golang:1.19-alpine AS base
RUN apk update && apk add git gcc musl-dev make
WORKDIR /app
COPY . .
RUN go mod download
RUN make build

FROM golang:1.19-alpine
WORKDIR /app
COPY --from=base /app/bin/bitheroes-guide-bot bitheroes-guide-bot
COPY --from=base /app/bin/data data
CMD ["/app/bitheroes-guide-bot"]
