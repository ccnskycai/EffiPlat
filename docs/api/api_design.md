*   **用户认证** (`/api/v1/auth`):
    *   `POST /login`
        *   **Request Body**: `{ "email": "user@example.com", "password": "your_password" }`
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Login successful", "data": { "token": "jwt_token_string", "userId": 1, "name": "John Doe", "roles": ["admin", "developer"] } }`
        *   **Response (Error - 401 Unauthorized)**: `{ "code": 40101, "message": "Invalid credentials", "data": null }`
    *   `POST /logout` (需要认证)
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Logout successful", "data": null }`
    *   `GET /me` (需要认证)
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "User info retrieved", "data": { "id": 1, "name": "John Doe", "email": "user@example.com" } }`
    *   详细设计见 [auth_api.md](./auth_api.md) 