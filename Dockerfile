FROM golang:1.19-alpine AS base
RUN apk update && apk add git gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /bitheroes-community-bot

FROM golang:1.19-alpine
COPY --from=base /bitheroes-community-bot /bitheroes-community-bot
CMD ["/bitheroes-community-bot"]
