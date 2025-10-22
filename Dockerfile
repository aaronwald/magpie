FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o magpie .

FROM golang:1.23
COPY --from=builder /app/magpie /app/magpie
CMD ["/app/magpie", "email", "magpie", "magpie", "mqtt",  "aaronwald@gmail.com", "aaronwald@gmail.com", "--gmail-username-file", "/etc/gmail-secret/username", \
  "--gmail-password-file", "/etc/gmail-secret/password", "--chough-addr", "chough:9099", \
  "--topics","mostert/#,+/events/rpc","prusa/basement/#"]
