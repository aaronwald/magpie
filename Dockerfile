FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o magpie .

FROM golang:1.23
COPY --from=builder /app/magpie /app/magpie
CMD ["/app/magpie"]
