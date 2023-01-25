FROM golang:1.19-alpine AS base
RUN apk update && apk add git gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bitheroes-guide-bot

FROM golang:1.19-alpine
WORKDIR /app
COPY --from=base /app/bitheroes-guide-bot bitheroes-guide-bot
COPY --from=base /app/data data
CMD ["/app/bitheroes-guide-bot"]
