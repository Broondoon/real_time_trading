FROM golang:1.23

WORKDIR /app

COPY server/ ./server
COPY Shared/ ./Shared

RUN go work init ./server
RUN go work use ./Shared

WORKDIR /app/server

RUN go mod tidy
RUN go build -o main .

CMD [ "./main" ]
