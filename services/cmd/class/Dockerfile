FROM golang:1.24-alpine3.21 as builder
LABEL authors="Pastor"

WORKDIR app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 go build -o /class -ldflags '-extldflags "-static"' pet/services/cmd/class

FROM scratch
COPY --from=builder /class /class

EXPOSE 51051
EXPOSE 8081

CMD ["/class"]