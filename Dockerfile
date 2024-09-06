
# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.22.7-alpine3.20 AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY go.mod ./

# download Go modules and dependencies
RUN go mod download

# copy directory files i.e all files ending with .go
COPY . .

# compile application
RUN go build -o /executor

FROM golang:1.22.7-alpine3.20 AS run

COPY --from=build ./executor ./executor

# Copy the config.yml file to the run stage
COPY --from=build /app/config.yml /config.yml

WORKDIR /
RUN go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
# tells Docker that the container listens on specified network ports at runtime
EXPOSE 8081

# command to be used to execute when the image is used to start a container
CMD [ "/executor" ]