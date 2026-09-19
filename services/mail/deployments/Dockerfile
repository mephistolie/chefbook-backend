FROM golang:1.26.2 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /mail ./cmd/app
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /mail /mail
ENTRYPOINT ["/mail"]
