# RateLimiter Service

## Overview
This repository contains the source code for the **RateLimiter** service, a Go-based application designed to control and limit the number of requests a client can make within a specified time frame. This service uses a gRPC API, ensuring scalability and efficiency for rate limiting in distributed systems. The project is containerized using Docker and is Kubernetes-ready for easy deployment and scalability.

## Problem Statement
The **RateLimiter** service's primary function is to provide a configurable rate limiting mechanism for clients, managing request allowances based on set limits (e.g., 100 requests per minute). 

### Key Features
- **Rate Limiting:** A gRPC API endpoint to handle rate limiting, with configurable limits.
- **Persistent Configuration:** Client-specific rate limits stored in PostgreSQL, with efficient fetching and caching.
- **Metrics & Logging:** Using tools like Prometheus for monitoring the number of processed and denied requests.
- **Docker & Kubernetes:** Containerization for easy deployment and scaling.

## Repository Structure

This section provides a detailed overview of the repository's structure, explaining the purpose of each file and directory.

- **`rate-limit/`**: Main directory containing all the service components and configurations.
  - **`amazoneks/`**: Configurations for Amazon EKS (Elastic Kubernetes Service).
    - `grpc-client-deployment.yaml`: Deployment configuration for the gRPC client.
    - `grpc-client-service.yaml`: Service configuration for the gRPC client.
    - `postgre-sql-operations-deployment.yaml`: Deployment config for PostgreSQL operations.
    - `postgre-sql-operations-service.yaml`: Service config for PostgreSQL operations.
    - `prometheus-configmap.yaml`: ConfigMap for Prometheus monitoring setup.
    - `prometheus-deployment.yaml`: Deployment configuration for Prometheus.
    - `prometheus-service.yaml`: Service configuration for Prometheus.
    - `prometheus.yml`: Main configuration file for Prometheus.
    - `rate-limiter-deployment.yaml`: Deployment configuration for the RateLimiter service.
    - `rate-limiter-env.yaml`: Environment configurations for RateLimiter.
    - `rate-limiter-service.yaml`: Service configuration for RateLimiter.
    - `redis-refresher-deployment.yaml`: Deployment config for Redis Refresher.
    - `redis-refresher-service.yaml`: Service config for Redis Refresher.
  - **`db/`**: Contains database-related files, such as schemas and migration scripts.
  - **`grpcClient/`**: Source code for the gRPC client.
    - `Dockerfile-grpcClient`: Dockerfile for building the gRPC client image.
    - `grpc_client.go`: Main Go source file for the gRPC client.
  - **`PostgreSQLOperations/`**: Code related to PostgreSQL operations.
  - **`Prometheus/`**:
    - `prometheus.yml`: Prometheus configuration duplicated from `amazoneks/`.
  - **`RateLimiter/`**: Source code for the RateLimiter service.
    - `Dockerfile-RateLimiter`: Dockerfile for the RateLimiter service.
    - `rateLimiter.go`: Main Go source file for the RateLimiter.
  - **`RedisRefresher/`**: Source code for the Redis Refresher component.
    - `Dockerfile-RedisRefresher`: Dockerfile for Redis Refresher.
    - `redisRefresher.go`: Go source file for Redis Refresher.

- **`docker-compose.yaml`**: Docker Compose configuration for local deployment and testing.

- **`go.mod`** & **`go.sum`**: Go module files managing project dependencies.

- **`Makefile`**: Contains commands for building, running, and managing the project.

- **`sqlc.yaml`**: Configuration file for SQLC, a tool for generating type-safe SQL queries.

# Rate-Limit

This project contains a comprehensive implementation of a rate-limiting service in Go, including several key functionalities. Let's break down the major components of the `rateLimiter.go`, 'grpc_client.go','postgresql_operation.go' and 'redisRefresher.go' files:

## `rateLimiter.go` - Rate Limiting Logic and gRPC Server 

The `rateLimiter.go` file in your project is a comprehensive implementation of a rate-limiting service in Go, involving several key functionalities. Let's break down its major components:

1. **Package and Imports**
   - The file is part of the main package.
   - Imports include standard libraries for networking, context management, and error handling.
   - External libraries like go-redis, lib/pq for PostgreSQL, gopsutil for system metrics, and grpc for the gRPC server are used.

2. **Global Variables**
   - Addresses for the gRPC server, Redis, and PostgreSQL connection strings.
   - Flags for dynamic rate limiting and system resource usage.

3. **RateLimitServer Struct**
   - Defines a struct RateLimitServer implementing proto.UnimplementedRateLimitServiceServer.
   - Contains Redis client and database connection for handling rate limits.

4. **NewRateLimitServer Function**
   - Constructor function for RateLimitServer, initializing it with Redis and DB clients.

5. **CheckRateLimit Method**
   - Implements the rate-limiting logic.
   - Checks if a request is allowed based on client ID and required tokens.
   - Logs the decision and returns the result.

6. **Redis-based Token Bucket Implementation**
   - Functions like initializeRateLimiter, refillTokens, and allowRequest manage a token bucket algorithm using Redis.
   - Handles token initialization, refilling, and decrementing based on rate limit configurations.

7. **PostgreSQL Integration**
   - fetchRateLimitConfigFromPostgres retrieves rate limit configurations from a PostgreSQL database.

8. **Dynamic Rate Limiting**
   - Based on system load (CPU and memory usage), calculated by getSystemLoad.
   - calculateNewLimits adjusts rate limits dynamically based on system resource usage.

9. **gRPC Server Setup**
   - startGRPCServer sets up and starts the gRPC server for handling incoming requests.
   - Registers the RateLimitServer with the gRPC server.

10. **Main Function**
    - Initializes dynamic rate limiting, Redis client, and PostgreSQL database connection.
    - Regularly updates system load metrics.
    - Starts the gRPC server and keeps the application running.

11. **Utility Functions**
    - pingServer and pingPostgresDB for health checks of Redis and PostgreSQL.



## `grpcClient.go` - gRPC Client for RateLimiter Service

This Go file, presumably named `grpcClient.go`, is designed to act as a gRPC client for the RateLimiter service. Here's a detailed breakdown of its functionalities and key components:

1. **Package and Imports**
   - Part of the main package.
   - Imports standard libraries for network communication, encoding, and logging.
   - External libraries include grpc for gRPC communication, Prometheus for metrics, and the project-specific proto package for RateLimit service definitions.

2. **Global Variables for Prometheus Metrics**
   - Defines Prometheus counters: allowedRequests and rejectedRequests to track the total number of allowed and rejected requests, respectively.

3. **Main Function**
   - Initializes Prometheus metrics.
   - Sets up an HTTP server to expose metrics and handle rate limit checks.
   - Establishes a connection to the gRPC server.
   - Defines an HTTP handler /check-rate-limit to process rate limit check requests.

4. **Prometheus Initialization**
   - Prometheusinit function registers the defined metrics with Prometheus.

5. **HTTP Server Setup**
   - startHTTPServer starts an HTTP server primarily for exposing Prometheus metrics.
   - Listens on the specified port and serves the /metrics endpoint.

6. **Rate Limit Check Handler**
   - handleCheckRateLimit handles HTTP POST requests, decodes the JSON body to extract the client ID, and then invokes the rate limit check via gRPC.
   - Updates Prometheus metrics based on the response (allowed or rejected).
   - Sends back a JSON response with the outcome and remaining tokens.

7. **gRPC Client Interaction**
   - CheckRateLimit function calls the CheckRateLimit method on the gRPC server.
   - Passes the client ID and a token requirement (hardcoded as 1 in this example).
   - Updates Prometheus metrics based on the gRPC response.

8. **Key Features**
   - Integration with a gRPC server to check rate limits.
   - Prometheus metrics for monitoring allowed and rejected requests.
   - An HTTP server to expose these metrics and handle rate limit check requests.
   - JSON request and response handling for web-based interactions.

9. **Usage**
   - Primarily acts as an intermediary between a web interface and the gRPC RateLimiter service.
   - Can be used for monitoring and controlling access to a system based on predefined rate limits.

This file complements the RateLimiter service by providing a client-side interface for rate limit checks, along with metrics monitoring capabilities. It demonstrates a practical application of gRPC client-server communication, enriched with Prometheus for real-time metrics tracking.


### Key Takeaways

- The file demonstrates a well-rounded approach to rate limiting, integrating Redis for token management and PostgreSQL for configuration storage.
- Dynamic rate limiting based on system load is a unique feature, adapting to changing server conditions.
- The use of gRPC enables efficient and scalable client-server communication.

This file is a critical component of your rate limiting service, encapsulating the core logic and server setup. It effectively combines database management, caching, and network communication to enforce rate limits for clients.


## `PostgreSQLOperations.go` - PostgreSQL Operations for Rate Limit Management

This Go file, presumably named `PostgreSQLOperations.go`, is focused on handling HTTP requests related to client rate limit management in a PostgreSQL database. Here's a detailed explanation of its functionalities:

1. **Package and Imports**
   - Part of the main package.
   - Imports include standard libraries for HTTP communication, database interaction, context management, and logging.
   - External libraries like pq for PostgreSQL handling, and a custom db package for database operations.

2. **Structs and Handlers**
   - CreateNewUserParams, DeleteClientParams, ListClientParams, GenerateRandomClientParams, UpdateClientParametersParams, and RateLimitParams are structs for handling different types of client requests.
   - HTTPHandler struct holds a reference to db.Queries, facilitating database operations.

3. **HTTP Handlers**
   - InsertNewClientHandler, DeleteClientHandler, ListClientIDRecordsHandler, GenerateRandomClientIDsHandler, UpdateClientParametersHandler, and DeleteAllClientsHandler are methods on HTTPHandler, each handling specific types of HTTP requests.

4. **Database Operations**
   - InsertNewClient, DeleteClient, DeleteAllClients, and UpdateClientParameters are functions that perform specific database operations like inserting, deleting, and updating client rate limit records.
   - These functions utilize the db.Queries methods to interact with the PostgreSQL database.

5. **HTTP Server Setup and Routing**
   - Sets up an HTTP server to listen on a specified port.
   - Routes like /insert, /delete, /list-clients, /generate-random-clients, /update-client-parameters, and /delete-all-clients are mapped to their respective handlers.

6. **Error Handling and Logging**
   - Each function and HTTP handler includes error handling and logging to capture and report any issues during execution.

### Key Features

- A comprehensive set of HTTP endpoints for managing rate limits in a PostgreSQL database.
- Structured and organized handling of HTTP requests and responses.
- Direct integration with a PostgreSQL database using custom query functions.

### Usage

- This file serves as a backend server for a web-based interface, where clients can manage their rate limit settings.
- It can be used for operations like adding new clients, updating rate limits, listing clients, and cleaning up the database.
- This file is essential for managing the rate limiting configurations of clients in the PostgreSQL database. It demonstrates a practical implementation of HTTP server setup, request handling, and database operations in Go.


## `RedisRefresher.go` - Redis Cache Refresh for Rate Limit Data

This Go file, likely named `RedisRefresher.go`, is designed to provide HTTP endpoints for refreshing Redis cache entries based on PostgreSQL database data. Here's a detailed breakdown of its key components and functionalities:

1. **Package and Imports**
   - Part of the main package.
   - Imports include standard libraries for HTTP communication, context management, JSON encoding, and logging.
   - External libraries include redis for Redis client operations and a custom db package for database interactions.

2. **Global Variables**
   - `redisClient`: A Redis client instance for interacting with the Redis cache.
   - `queries`: An instance of `db.Queries` for executing database operations.

3. **Main Function**
   - Initializes database and Redis connections using predefined connection strings.
   - Sets up two HTTP handlers, `/refresh/id` and `/refresh/all`, for individual and bulk cache refresh operations.

4. **HTTP Server Setup and Handlers**
   - Starts an HTTP server listening on a specified port.
   - `handleRefreshID` handles requests for refreshing the Redis cache for a specific client ID.
   - `handleRefreshAll` handles requests for refreshing the entire Redis cache.

5. **Cache Refresh Operations**
   - `refreshID` function retrieves the current rate limit data for a given client ID from the PostgreSQL database and updates the corresponding entries in Redis.
   - `refreshAll` function clears the entire Redis cache.

6. **Error Handling and Response Functions**
   - `respondWithError` and `respondWithJSON` are utility functions to send error messages and JSON responses to HTTP requests, respectively.

### Key Features:

- Provides endpoints for refreshing rate limit data in Redis cache based on the latest values from a PostgreSQL database.
- Supports both individual record refresh (based on client ID) and bulk refresh of the entire cache.
- Direct interaction with Redis and a PostgreSQL database through respective clients.

### Usage:

- This file can be used in scenarios where the rate limit data in the Redis cache needs to be updated or reset, either for specific clients or entirely.
- It serves as a maintenance or administrative tool within the rate limiting system, ensuring that the cache is consistent with the database records.
- This file is crucial for maintaining the integrity and accuracy of the rate limiting system, especially in cases where rate limits are updated or need to be reset in the cache. It demonstrates a practical implementation of HTTP server functionalities, database queries, and Redis cache operations in Go.
## Getting Started of Usage RateLimiting Service

### Prerequisites
- Postman

# Rate Limit Postman Collection Usage Guide

This documentation explains how to use the "Rate Limit" Postman Collection. This collection contains various HTTP requests to a specific API, including database operations, Redis cache management, and rate limit checking.

## Setup

Ensure you have Postman installed on your computer and have a Postman account. You can then import this collection into Postman.

1. Open the Postman application.
2. Click on the "Import" button.
3. Copy the JSON content from the top of this document and paste it into Postman, or directly drag and drop the JSON file into Postman.

## Collection Contents

### 1. Database Operations

#### 1.1. Deleting ClientID

- Used to delete a specific `ClientID` from the database.

#### 1.2. Inserting ClientID

- Used to add a new `ClientID`. This operation includes the rate limit and refill interval.

#### 1.3. List All ClientIDs

- Used to list all `ClientIDs` in the database.

#### 1.4. Generate Random ClientIDs

- Generates random `ClientIDs`.

#### 1.5. Update ClientID Parameters

- Updates the rate limit and refill interval for a `ClientID`.

#### 1.6. Delete All Database

- Deletes all records in the database.

### 2. Refresh Redis Cache

#### 2.1. Refresh with Client ID Number

- Refreshes the cache with a specific `ClientID`.

#### 2.2. Refresh All Cache

- Refreshes the entire cache.

### 3. Check Rate Limit

#### 3.1. Check Rate Limit with ClientID

- Checks the rate limit for a specific `ClientID`.





