FROM golang:1.23

WORKDIR /app

COPY order-executor-service/ ./order-executor-service
COPY Shared/ ./Shared

RUN go work init ./order-executor-service
RUN go work use ./Shared

WORKDIR /app/order-executor-service

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
