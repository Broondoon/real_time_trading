FROM golang:1.23

WORKDIR /app

COPY client/ ./client
COPY Shared/ ./Shared

RUN go work init ./client
RUN go work use ./Shared

WORKDIR /app/client

RUN go mod tidy
RUN go build -o main .

CMD [ "./main" ]
