# API Documentation

only a stub...

## Chirp routes

### GET /api/chirps

Returns all chirps.

#### URL

`GET /api/chirps`

#### Path Parameters

_None_

#### Query Parameters

- `author_id` (`string(uuid)`): If added will only return Chirps of this user

#### Request Headers

- `Authorization`: `Bearer <token>` (required)

#### Responses

##### 201 CREATED

```json
[{
    "id": "123",
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z",
    "body": "chirp text",
    "user_id": "123"
}]
```

##### 401 Unauthorized

```json
{
    "error": "Unauthorized"
}
```

##### 500 Internal Server Error

```json
{
    "error": "Internal Server Error"
}
```

##### 400 BadRequest

```json
{
    "error": "Chirp is too long"
}
```
