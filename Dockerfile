FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bookmart-api .
FROM alpine:3.20
RUN adduser -D -H appuser
COPY --from=build /bookmart-api /usr/local/bin/bookmart-api
USER appuser
EXPOSE 8080
CMD ["bookmart-api"]
