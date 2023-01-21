FROM golang:1.19-alpine AS base
RUN apk update && apk add git gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bitheroes-community-bot

FROM golang:1.19-alpine
WORKDIR /app
COPY --from=base /app/bitheroes-community-bot bitheroes-community-bot
COPY --from=base /app/data data
CMD ["/app/bitheroes-community-bot"]
