# gRPC Server

gRPC Server is a Go application that integrates a gRPC (Google Remote Procedure Calls) server, Kafka messaging, Jaeger distributed tracing, and the `go.uber.org/zap` logger.

## Table of Contents
- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
  - [Get](#get)
  - [Create](#create)
  - [Delete](#delete)
  - [Update](#update)
- [Distributed Tracing with Jaeger](#distributed-tracing-with-jaeger)
- [Testing](#testing)

## Introduction

The Go-based GRPC Server is a versatile application that offers a gRPC server, Kafka messaging, Jaeger distributed tracing, and the zgo.uber.org/zap logger. This solution empowers developers to construct scalable and high-performance server applications using the gRPC, while concurrently supporting REST API functionality. By incorporating Kafka messaging, the server facilitates asynchronous communication and enhances component decoupling. Additionally, the integration of Jaeger distributed tracing aids in monitoring and pinpointing performance issues within distributed systems. The go.uber.org/zap logger ensures efficient logging, minimizing performance impact.
## Prerequisites

Before running this application, ensure that you have the following prerequisites installed:

- Go: [Install Go](https://go.dev/doc/install/)
- Docker: [Install Docker](https://docs.docker.com/get-docker/)
- Docker Compose: [Install Docker Compose](https://docs.docker.com/compose/install/)

## Installation

1. Clone this repository
  ```bash
    git clone https://github.com/NRKA/gRPC-Server.git
  ```
2. Navigate to the project directory:
  ```
    cd gRPC-Server
  ```
3. Build the Docker image:
  ```
    docker-compose build
  ```

## Usage
1. Start the Docker containers:
  ```
    docker-compose up
  ```
2. The application will be accessible at:
  ```
    localhost:9000/
  ```

## API Endpoints

### Get

- Method: GET
- Endpoint: /entity
- Query Parameter: id=[entity_id]

Retrieve data from the database based on the provided ID.

- Response:
  - 200 OK: Returns the data in the response body.
  - 400 Bad Request: If the `id` query parameter is missing.
  - 404 Not Found: If the provided ID does not exist in the database.
  - 500 Internal Server Error: If there is an internal server error.

### Create

- Method: POST
- Endpoint: /entity
- Request Body: JSON payload containing the ID and data.

Add new data to the database.

- Response:
  - 200 OK: If the request is successful.
  - 500 Internal Server Error: If there is an internal server error.

### Delete

- Method: DELETE
- Endpoint: /entity
- Query Parameter: id=[entity_id]

Remove data from the database based on the provided ID.

- Response:
  - 200 OK: If the request is successful.
  - 404 Not Found: If the provided ID does not exist in the database.
  - 500 Internal Server Error: If there is an internal server error.

### Update

- Method: PUT
- Endpoint: /entity
- Request Body: JSON payload containing the ID and updated data.

Update existing data in the database based on the provided ID.

- Response:
  - 200 OK: If the request is successful.
  - 404 Not Found: If the provided ID does not exist in the database.
  - 500 Internal Server Error: If there is an internal server error.

## Distributed Tracing with Jaeger

This project integrates with Jaeger for distributed tracing. Jaeger allows you to trace the flow of requests across multiple services, providing insights into performance and identifying bottlenecks in the system.

To view the traces, access the Jaeger UI at:
```
http://localhost:16686/
```
For more information on how to use Jaeger for distributed tracing in Go, refer to the Jaeger Go client documentation.

## Testing

- To run unit tests, use the following command:
  ```bash
    go test ./... -cover
  ```
- To run integration tests, use the following command:
  ```bash
    go test -tags integration ./... -cover
  ```
