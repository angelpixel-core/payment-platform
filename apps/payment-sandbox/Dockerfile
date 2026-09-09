FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/payment-sandbox ./cmd/payment-sandbox

FROM alpine:3.22

RUN adduser -D -H app
COPY --from=build /out/payment-sandbox /usr/local/bin/payment-sandbox
COPY docs /app/docs
WORKDIR /app
USER app
EXPOSE 8080
ENTRYPOINT ["payment-sandbox"]
