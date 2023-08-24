FROM golang:1.21.0-alpine

# Set the Current Working Directory inside the container
WORKDIR /app 

#copy files
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

#build
RUN go build -o api .

# Expose port 8000 to the outside world
EXPOSE 8000

# run exe
CMD ["./api"]