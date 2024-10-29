# Microservices Demo with Circuit Breaker

This project demonstrates a microservices architecture with two services (Service A and Service B) implementing a circuit breaker pattern using Go, Redis, and Docker.

## Project Structure

- `serviceA/`: Contains the code for Service A
- `serviceB/`: Contains the code for Service B
- `docker-compose.yml`: Defines the multi-container Docker environment

## Services

### Service A

Service A is the main service that calls Service B and implements the circuit breaker pattern.

Key features:
- Uses the `gobreaker` library for circuit breaking
- Integrates with Redis for distributed circuit state management
- Exposes an HTTP endpoint `/call-service-b`

Dependencies:
```go
require (
	github.com/go-kit/kit v0.13.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/sony/gobreaker v0.5.0
)
```

### Service B

Service B is a simple service that simulates occasional failures.

Key features:
- Exposes an HTTP endpoint `/hello`
- Simulates failures every third request

## Circuit Breaker Pattern

The circuit breaker is implemented in Service A using the following components:

1. `gobreaker` library for local circuit breaking
2. Redis for distributed circuit state management

Circuit breaker settings:
```go
var st gobreaker.Settings
st.Name = "ServiceB"
st.MaxRequests = 3
st.Interval = time.Duration(5) * time.Second
st.Timeout = time.Duration(10) * time.Second
st.ReadyToTrip = func(counts gobreaker.Counts) bool {
    failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
    return counts.Requests >= 3 && failureRatio >= 0.6
}
```

## Redis Integration

Redis is used to store the circuit state, allowing multiple instances of Service A to share the same circuit state.

## Docker Setup

The project uses Docker Compose to set up and run all services. The `docker-compose.yml` file defines the following services:

1. Redis
2. Service B
3. Two instances of Service A

To build and run the project:

```bash
docker-compose up --build
```

## API Endpoints

- Service A: `http://localhost:8080/call-service-b` and `http://localhost:8082/call-service-b`
- Service B: `http://localhost:8081/hello`

## Testing the Circuit Breaker

To test the circuit breaker:

1. Send multiple requests to Service A's `/call-service-b` endpoint
2. Observe how Service A handles failures from Service B
3. When the circuit opens, Service A will return a 503 Service Unavailable response
4. The circuit will automatically close after the specified timeout

## Future Improvements

- Add unit tests for both services
- Implement retry mechanisms
- Add monitoring and logging for better observability
- Implement a fallback mechanism when the circuit is open
