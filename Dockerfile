FROM golang:1.25 AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o app

FROM gcr.io/distroless/static-debian13:latest
COPY --from=builder /build/app /
CMD [ "/app" ]
