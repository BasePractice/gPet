FROM golang:1.24-alpine3.21
LABEL authors="Pastor"

WORKDIR app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .
RUN go build -o /class pet/services/cmd/class
EXPOSE 51051

CMD ["/class"]