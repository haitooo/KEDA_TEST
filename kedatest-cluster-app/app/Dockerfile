FROM golang:1.24 as builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/app /app
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app"]
