FROM golang:1.23 AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o app

FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /build/app /
CMD [ "/app" ]
