FROM golang:1.24-alpine3.21 as builder
LABEL authors="Pastor"

WORKDIR app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 go build -o /hasq -ldflags '-extldflags "-static"' pet/services/cmd/hasq

FROM scratch
COPY --from=builder /hasq /hasq

EXPOSE 52051
EXPOSE 8081

CMD ["/hasq"]