# Employee Management System API

This is a Golang-based RESTful API for managing employee records, built with the Echo web framework, PostgreSQL for persistent storage, Redis for caching, and Swagger for API documentation. The API supports CRUD operations for employee records, JWT-based authentication for secured endpoints, and containerization with Docker.

## Features
- **CRUD Endpoints**: Create, retrieve, update, and delete employee records.
- **JWT Authentication**: Secured endpoints (`POST /employees`, `PUT /employees/{id}`, `DELETE /employees/{id}`) require an `Authorization: Bearer <token>` header.
- **Database**: PostgreSQL with `pgx` driver for raw SQL queries (no ORM).
- **Caching**: Redis caching for `GET /employees` and `GET /employees/{id}` to improve read performance.
- **Swagger Documentation**: Interactive API documentation via Swagger UI at `/swagger/*`.
- **Error Handling**: Consistent error responses with appropriate HTTP status codes.
- **Testing**: Unit and integration tests for controllers (see `tests/controller_test.go`).
- **Docker Support**: Containerized application for easy deployment.

## Project Structure
```
├── cmd
│   └── main.go               # Application entry point
├── config
│   └── config.go             # Configuration loading (environment variables)
├── controller
│   └── controller.go         # HTTP handlers with Swagger annotations
├── customerr
│   └── err.go                # Custom error handling
├── database
│   ├── model.go              # Data models (Employee, Credentials, etc.)
│   ├── psql.go               # PostgreSQL connection setup
│   └── redis.go              # Redis connection setup
├── Dockerfile                # Docker configuration
├── docs
│   ├── docs.go               # Generated Swagger documentation
│   ├── swagger.json          # Generated Swagger JSON
│   └── swagger.yaml          # Generated Swagger YAML
├── employee.sql              # SQL queries for employee operations
├── go.mod                    # Go module dependencies
├── go.sum                    # Go module checksums
├── middleware
│   └── middleware.go         # JWT authentication and logging middleware
├── repo
│   ├── db.go                 # Database interface
│   ├── employee.sql.go       # SQLC-generated database code
│   ├── models.go             # SQLC-generated models
│   └── repo.go               # Repository layer for database operations
├── routes
│   └── route.go              # API route definitions
├── schema.sql                # Database schema for employees table
├── service
│   └── service.go            # Business logic layer
├── sqlc.yaml                 # SQLC configuration
├── tests
│   └── controller_test.go    # Unit and integration tests
└── tmp
    ├── build-errors.log      # Build error logs
    └── main                  # Temporary build output
```

## Prerequisites
- **Go**: Version 1.20 or higher
- **PostgreSQL**: Version 12 or higher
- **Redis**: Version 6 or higher
- **Docker**: (Optional) For containerized deployment
- **swag**: For generating Swagger documentation
- **sqlc**: For generating database code from SQL queries

## Setup Instructions
### 1. Clone the Repository
```bash
git clone https://github.com/lijuuu/EmployeeManagement.git
cd EmployeeManagement
```

### 2. Install Dependencies
Install Go dependencies:
```bash
go mod tidy
```

Install `swag` for Swagger documentation:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Install `sqlc` for database code generation:
```bash
go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
```

### 3. Set Up PostgreSQL
1. Ensure PostgreSQL is running locally or in a Docker container:
   ```bash
   docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=yourpassword postgres
   ```
2. Create the database and employees table:
   ```bash
   psql -h localhost -U postgres -d postgres -f schema.sql
   ```
   The `schema.sql` file defines the `employees` table:
   ```sql
   CREATE TABLE employees (
       id SERIAL PRIMARY KEY,
       name VARCHAR NOT NULL,
       position VARCHAR NOT NULL,
       salary INTEGER NOT NULL,
       hired_date DATE NOT NULL,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

### 4. Set Up Redis
1. Ensure Redis is running locally or in a Docker container:
   ```bash
   docker run -d --name redis -p 6379:6379 redis
   ```

### 5. Configure Environment Variables
Create a `.env` file in the project root with the following:
```env
POSTGRES_DSN=postgresql://postgres:yourpassword@localhost:5432/postgres?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=securepassword
JWT_SECRET=your-jwt-secret
```
Replace `yourpassword` and `your-jwt-secret` with secure values.

### 6. Generate Database Code
Generate database code using `sqlc`:
```bash
sqlc generate
```
This processes `employee.sql` and `sqlc.yaml` to generate `repo/employee.sql.go` and `repo/models.go`.

### 7. Generate Swagger Documentation
Generate Swagger files (`docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`):
```bash
swag init -g cmd/main.go -o docs
```

### 8. Run the Application
Start the API server:
```bash
go run cmd/main.go
```
The API will be available at `http://localhost:8080`.

### 9. Access Swagger UI
Open `http://localhost:8080/swagger/index.html` to view the interactive API documentation.

## Using the API
### Authentication
1. **Login** to obtain a JWT token:
   ```bash
   curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"securepassword"}'
   ```
   Response:
   ```json
   {"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
   ```

2. **Use the Token** for secured endpoints (`POST /employees`, `PUT /employees/{id}`, `DELETE /employees/{id}`):
   - Include the `Authorization: Bearer <token>` header.
   - Example:
     ```bash
     curl -X POST http://localhost:8080/employees \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <your_jwt_token>" \
       -d '{"name":"John Doe","position":"Software Engineer","salary":60000,"hired_date":"2024-06-01"}'
     ```
     Response:
     ```json
     {
       "id":"1",
       "name":"John Doe",
       "position":"Software Engineer",
       "salary":60000,
       "hired_date":"2024-06-01",
       "created_at":"2024-06-10T12:00:00Z",
       "updated_at":"2024-06-10T12:00:00Z"
     }
     ```

### Endpoints
- **POST /login**: Authenticate admin and return a JWT token.
- **POST /employees**: Create a new employee (requires JWT).
- **GET /employees**: List all employees (cached in Redis).
- **GET /employees/{id}**: Retrieve an employee by ID (cached in Redis).
- **PUT /employees/{id}**: Update an employee (requires JWT).
- **DELETE /employees/{id}**: Delete an employee (requires JWT).

### Swagger UI
- Access: `http://localhost:8080/swagger/index.html`
- Authorize: Click the "Authorize" button, enter `Bearer <token>` (e.g., `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`).
- Test endpoints interactively, ensuring the `Authorization` header is sent for secured endpoints.

## Running with Docker
1. Build the Docker image:
   ```bash
   docker build -t employee-management .
   ```
2. Run the container, linking to PostgreSQL and Redis:
   ```bash
   docker run -d --name employee-api \
     -p 8080:8080 \
     --link postgres:postgres \
     --link redis:redis \
     -e POSTGRES_DSN=postgresql://postgres:yourpassword@postgres:5432/postgres?sslmode=disable \
     -e REDIS_ADDR=redis:6379 \
     -e REDIS_PASSWORD= \
     -e REDIS_DB=0 \
     -e ADMIN_EMAIL=admin@example.com \
     -e ADMIN_PASSWORD=securepassword \
     -e JWT_SECRET=your-jwt-secret \
     employee-management
   ```

## Testing
Run unit and integration tests:
```bash
go test ./tests/...
```
The `tests/controller_test.go` file includes tests for all endpoints, including JWT-protected routes.


## License
MIT License. See [LICENSE](LICENSE) for details.
