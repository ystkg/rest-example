FROM golang:1.22 as build
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o app

FROM gcr.io/distroless/static-debian12:latest
COPY --from=build /build/app /
CMD [ "/app" ]
