FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o magpie .

FROM golang:1.23
COPY --from=builder /app/magpie /app/magpie
CMD ["/app/magpie", "email", "magpie", "magpie", "mqtt", "mostert/#",
    "aaronwald@gmail.com", "aaronwald@gmail.com",
    "--gmail-username-file", "/etc/gmail-secret/username",
    "--gmail-password-file", "/etc/gmail-secret/password"]
#CMD ["tail", "-f", "/dev/null"]
