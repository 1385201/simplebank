
#构建阶段
FROM golang:1.25.0-alpine AS builder
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=$GOPROXY
WORKDIR /app
COPY . .
RUN go build -o main main.go



#运行阶段
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .

EXPOSE 8080
CMD ["/app/main"]