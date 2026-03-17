FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o bike-rental ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache sqlite-libs ca-certificates

WORKDIR /app
COPY --from=builder /app/bike-rental .

ENV PORT=8080
ENV DB_PATH=/data/bike_rental.db
ENV JWT_SECRET=change-me-in-production
ENV ADMIN_CREDENTIALS=YWRtaW46cGFzc3dvcmQ=

VOLUME ["/data"]
EXPOSE 8080

CMD ["./bike-rental"]
