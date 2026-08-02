# Fitness API
A simple API to track gym workouts, built with Go's standard library. Developed using [REST Servers in GO](https://eli.thegreenplace.net/2021/rest-servers-in-go-part-1-standard-library/) reference to learn building APIs in Go.

## Features
- Creates and retrieves exercises
- Structured request logging
- Rate limiting middleware
- HTTPS support using self signed TLS certificate
- Token Based Authentication
- Panic recovery 

## Setup
Clone the repository and start the server:
```bash
SERVERPORT=4112 go run .
```
Replace ```4112``` with the port you want the API to use.

## Example Request
```bash
curl --cacert cert.pem \
    -H "Content-Type: application/json" \
    --data '{
        "name": "Bench Press",
        "sets": 3,
        "reps": 5,
        "weight": 160,
        "tags": ["push", "chest"]
    }' \
    https://localhost:4112/exercise/
```

## Project Notes
This project is used as a learning tool for backend development in Go. No AI was used in the development of this API to maximize learning opportunity. 

## Future Features
- Relate exercise store items to specific users. Add authentication logic based on user ownership over exercise entries.
- Add expiration token logic and www-authenticate headers.
- Create admin authentication to lock specific endpoints behind admin privelages. 
- Design and add pagination and scalability features