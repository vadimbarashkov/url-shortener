# URL Shortener API Documentation

## Overview

The URL Shortener API allows users to create, manage, and retrieve shortened URLs. This documentation provides details on how to use the API, including available endpoints, request/response formats, error codes, and general usage guidelines.

### Base URL

All API requests are prefixed with:

```bash
http://localhost:8080/api/v1
https://localhost:8443/api/v1
```

It depends on the environment in which the project is launched.

### Supported Content-Type

- `application/json`

## Authentication

There are no authentication requirements for this version of the API. All endpoints are publicly accessible.

---

## Endpoints

### 1. Health Check - `GET /ping`

**Description:**
A simple health check to verify that the server is running.

**Request:**

- **Method:** `GET`
- **Endpoint:** `/ping`
- **Body:** None

**Response:**

- **Status Code:** `200 OK`
- **Body:**

    ```json
    pong
    ```

---

### 2. Shorten a URL - `POST /shorten`

**Description:**
Shorten a given URL and return a unique short code that can be used to access the original URL.

**Request:**

- **Method:** `POST`
- **Endpoint:** `/shorten`
- **Body:**

    ```json
    {
        "url": "https://example.com"
    }
    ```

  - `url` (string, required): The original URL to shorten. Must be a valid URL.

**Response:**

- **Status Code:** `201 Created`
- **Body:**

    ```json
    {
        "status": "success",
        "message": "The URL has been shortened successfully.",
        "data": {
            "id": 1,
            "short_code": "abc123",
            "url": "https://example.com",
            "created_at": "2024-09-12T12:34:56Z",
            "updated_at": "2024-09-12T12:34:56Z"
        }
    }
    ```

**Errors:**

- **400 Bad Request:**

    ```json
    {
        "status": "error",
        "message": "Request body is empty. Please provide necessary data."
    }
    ```

    ```json
    {
        "status": "error",
        "message": "Invalid request body. Please check your input.",
        "details": [
            {
                "field": "url",
                "value": "https://example",
                "issue": "Invalid url."
            }
        ]
    }
    ```

- **500 Internal Server Error:**

    ```json
    {
        "status": "error",
        "message": "An internal server error occurred. Please try again later."
    }
    ```

---

### 3. Resolve a Short Code - `GET /shorten/{shortCode}`

**Description:**
Retrieve the original URL associated with the provided short code.

**Request:**

- **Method:** `GET`
- **Endpoint:** `/shorten/{shortCode}`
- `shortCode` (string, required): The short code of the URL.

**Response:**

- **Status Code:** `200 OK`
- **Body:**

    ```json
    {
        "status": "success",
        "message": "The short code was successfully resolved.",
        "data": {
            "id": 1,
            "short_code": "abc123",
            "url": "https://example.com",
            "created_at": "2024-09-12T12:34:56Z",
            "updated_at": "2024-09-12T12:34:56Z"
        }
    }
    ```

**Errors:**

- **404 Not Found:**

    ```json
    {
        "status": "error",
        "message": "The requested resource was not found."
    }
    ```

- **500 Internal Server Error:**

    ```json
    {
        "status": "error",
        "message": "An internal server error occurred. Please try again later."
    }
    ```

---

### 4. Modify a Shortened URL - `PUT /shorten/{shortCode}`

**Description:**
Modify the original URL linked to the provided short code.

**Request:**

- **Method:** `PUT`
- **Endpoint:** `/shorten/{shortCode}`
- `shortCode` (string, required): The short code of the URL to modify.

- **Body:**

    ```json
    {
        "url": "https://new-example.com"
    }
    ```

**Response:**

- **Status Code:** `200 OK`
- **Body:**

    ```json
    {
        "status": "success",
        "message": "The URL was successfully modified.",
        "data": {
            "id": 1,
            "short_code": "abc123",
            "url": "https://new-example.com",
            "created_at": "2024-09-12T12:34:56Z",
            "updated_at": "2024-09-12T12:40:00Z"
        }
    }
    ```

**Errors:**

- **400 Bad Request:**

    ```json
    {
        "status": "error",
        "message": "Request body is empty. Please provide necessary data."
    }
    ```

    ```json
    {
        "status": "error",
        "message": "Invalid request body. Please check your input.",
        "details": [
            {
                "field": "url",
                "value": "https://example",
                "issue": "Invalid url."
            }
        ]
    }
    ```

- **404 Not Found:**

    ```json
    {
        "status": "error",
        "message": "The requested resource was not found."
    }
    ```

- **500 Internal Server Error:**

    ```json
    {
        "status": "error",
        "message": "An internal server error occurred. Please try again later."
    }
    ```

---

### 5. Deactivate a Shortened URL - `DELETE /shorten/{shortCode}`

**Description:**
Deactivate a shortened URL, making the short code inactive and the URL no longer resolvable.

**Request:**

- **Method:** `DELETE`
- **Endpoint:** `/shorten/{shortCode}`
- `shortCode` (string, required): The short code of the URL to deactivate.

**Response:**

- **Status Code:** `200 OK`
- **Body:**

    ```json
    {
        "status": "success",
        "message": "The URL was successfully deactivated."
    }
    ```

**Errors:**

- **404 Not Found:**

    ```json
    {
        "status": "error",
        "message": "The requested resource was not found."
    }
    ```

- **500 Internal Server Error:**

    ```json
    {
        "status": "error",
        "message": "An internal server error occurred. Please try again later."
    }
    ```

---

### 6. Get URL Statistics - `GET /shorten/{shortCode}/stats`

**Description:**
Retrieve statistics for the URL associated with the given short code, such as the access count.

**Request:**

- **Method:** `GET`
- **Endpoint:** `/shorten/{shortCode}/stats`
- `shortCode` (string, required): The short code of the URL.

**Response:**

- **Status Code:** `200 OK`
- **Body:**

    ```json
    {
        "status": "success",
        "message": "The URL statistics retrieved successfully.",
        "data": {
            "id": 1,
            "short_code": "abc123",
            "url": "https://example.com",
            "access_count": 15,
            "created_at": "2024-09-12T12:34:56Z",
            "updated_at": "2024-09-12T12:34:56Z"
        }
    }
    ```

**Errors:**

- **404 Not Found:**

    ```json
    {
        "status": "error",
        "message": "The requested resource was not found."
    }
    ```

- **500 Internal Server Error:**

    ```json
    {
        "status": "error",
        "message": "An internal server error occurred. Please try again later."
    }
    ```

---

## Error Handling

### Common Error Responses

- **400 Bad Request:** Empty request body.

    ```json
    {
        "status": "error",
        "message": "Request body is empty. Please provide necessary data."
    }
    ```

- **400 Bad Request:** Invalid input data or malformed request.

    ```json
    {
        "status": "error",
        "message": "Invalid request body. Please check your input.",
        "details": [
            {
                "field": "url",
                "value": "invalid-url",
                "issue": "Invalid url."
            }
        ]
    }
    ```

- **404 Not Found:** The requested resource does not exist.

    ```json
    {
        "status": "error",
        "message": "The requested resource was not found."
    }
    ```

- **500 Internal Server Error:** Something went wrong on the server.

    ```json
    {
        "status": "error",
        "message": "An internal server error occurred. Please try again later."
    }
    ```

---

## Rate Limiting & Caching

- **Rate Limiting:**
  Currently, there are no rate limits enforced. However, we recommend clients implement retries with exponential backoff in case of server overload.

- **Caching:**
  Responses are not cached. Short URL resolution happens in real-time.

---

## Best Practices

1. **Use URL Validation:** Always ensure the `url` field provided is valid and follows the `http://` or `https://` protocol.
2. **Error Handling:** Handle all error responses gracefully in your client applications by inspecting the `status` and `message` fields in the response.
3. **Request Rate Optimization:** Avoid unnecessary requests to the API by ensuring the correct URL is passed in `POST` and `PUT` requests.
4. **Pagination (Future Consideration):** For future enhancements, endpoints may support pagination for retrieving large datasets (e.g., statistics).

---

## Versioning

This is version 1 (`v1`) of the API. Future versions will be released as breaking changes or significant new features are introduced.

---

## Contact

For any issues, contact the API support team at `vadimdominik2005@gmail.com`.
