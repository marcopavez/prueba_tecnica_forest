# Bike Rental API

A RESTful API service for a bike rental company written in Go.

## Tech Stack

- **Language**: Go 1.21+
- **Router**: go-chi/chi v5
- **Database**: SQLite (via mattn/go-sqlite3)
- **Auth**: JWT (user) + Basic Auth (admin)

## Quick Start

### Environment Variables

| Variable           | Default                     | Description                        |
|--------------------|-----------------------------|------------------------------------|
| `PORT`             | `8080`                      | Server port                        |
| `DB_PATH`          | `./bike_rental.db`          | SQLite database file path          |
| `JWT_SECRET`       | `default-secret-for-testing`| HMAC SHA256 secret for JWT signing |
| `ADMIN_CREDENTIALS`| `YWRtaW46cGFzc3dvcmQ=`      | Base64 of `username:password`      |

Default admin credentials: `admin:password`

### Run Locally

```bash
go mod tidy
JWT_SECRET=mysecret ADMIN_CREDENTIALS=YWRtaW46cGFzc3dvcmQ= go run ./cmd/main.go
```

### Run with Docker

```bash
docker build -t bike-rental .
docker run -p 8080:8080 \
  -e JWT_SECRET=mysecret \
  -e ADMIN_CREDENTIALS=YWRtaW46cGFzc3dvcmQ= \
  -v $(pwd)/data:/data \
  bike-rental
```

## API Reference

### Utility

#### Health Check
```
GET /status
```
Response: `{"status":"ok"}`

---

### User Endpoints

#### Register
```
POST /users/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login
```
POST /users/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```
Returns: `{"token": "<jwt>"}`

#### Get Profile *(requires JWT)*
```
GET /users/profile
Authorization: Bearer <token>
```

#### Update Profile *(requires JWT)*
```
PATCH /users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Smith",
  "password": "newpassword"
}
```

---

### Bike Endpoints *(requires JWT)*

#### List Available Bikes
```
GET /bikes/available
Authorization: Bearer <token>
```

---

### Rental Endpoints *(requires JWT)*

#### Start Rental
```
POST /rentals/start
Authorization: Bearer <token>
Content-Type: application/json

{"bike_id": 1}
```

#### End Rental
```
POST /rentals/end
Authorization: Bearer <token>
```

#### Rental History
```
GET /rentals/history
Authorization: Bearer <token>
```

---

### Admin Endpoints *(requires Basic Auth)*

All admin endpoints require `Authorization: Basic <base64(username:password)>`.

#### Bikes
```
POST   /admin/bikes
GET    /admin/bikes
PATCH  /admin/bikes/{bike_id}
```

#### Users
```
GET    /admin/users
GET    /admin/users/{user_id}
PATCH  /admin/users/{user_id}
```

#### Rentals
```
GET    /admin/rentals
GET    /admin/rentals/{rental_id}
PATCH  /admin/rentals/{rental_id}
```

---

## Business Logic

- A bike that is currently rented cannot be rented by another user.
- A user can only rent one bike at a time.
- When a rental ends, the bike's location is updated to a random point within 5km of the start location.
- Rental duration is calculated in minutes (rounded up).
- Cost is calculated as `duration_minutes × price_per_minute`.

## Project Structure

```
.
├── cmd/
│   └── main.go               # Entry point, router setup
├── internal/
│   ├── auth/
│   │   ├── jwt.go            # JWT generation & validation
│   │   └── basic.go          # Basic Auth validation
│   ├── db/
│   │   └── db.go             # DB connection & migrations
│   ├── handlers/
│   │   ├── helpers.go        # JSON helpers
│   │   ├── user.go           # User endpoints
│   │   ├── bike.go           # Bike endpoints
│   │   ├── rental.go         # Rental endpoints
│   │   └── admin.go          # Admin endpoints
│   ├── middleware/
│   │   ├── user_auth.go      # JWT middleware
│   │   └── admin_auth.go     # Basic Auth middleware
│   └── models/
│       └── models.go         # Data models
├── Dockerfile
├── go.mod
└── README.md
```
