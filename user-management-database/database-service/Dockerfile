FROM golang:1.23

WORKDIR /app

COPY user-management-database/database-service ./databaseServiceUserManagement
COPY Shared/ ./Shared

RUN go work init ./databaseServiceUserManagement
RUN go work use ./Shared

WORKDIR /app/databaseServiceUserManagement

RUN go mod tidy 
RUN go build -o main .

CMD [ "./main"]
