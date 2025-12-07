# FAQ API Service (Golang)

A service that provides a REST API for managing questions (`Question`)
and answers (`Answer`), stores data in PostgreSQL and uses migrations via `goose`.

---

## Technologies

- Go
- GORM
- PostgreSQL
- Goose:
- Docker + docker-compose

---

## Quick start

Requires docker and docker-compose installed

```bash
git clone https://github.com/AYaSmyslov/faqapi 
cd faqapi
docker-compose up --build
```

After a succesful launch:
- PostgreSQL runs in a `db` container
- Migrations are applied by the `migtare` container (goose)
- The application starts in the `api` container
- The API available at: `http://localhost:8080`

---

## Envitonment variables
In `docker-compose.yml` main variables have already been set:
- `DB_HOST` - PostgreSQL host (`db` by default)
- `DB_PORT` - PostgreSQL posrt (`5432` by default)
- `DB_USER` - the database user (`postgres` by default)
- `DB_PASSWORD` - the database user password (`postgres` by default)
- `DB_NAME` - the name of database (`faq_db` by default)
- `DB_SSLMODE` - SSL mode (`disable` by default)
- `DB_TIMEZONE` - time zone (`UTC` by default)
- `HTTP_ADDR` - the address/port of the HTTP server (`:8080` by default)

---

## Migrations (goose)
Migrations are stored in the `migrations/` directory and are automatically applies
by the `migrate` container when starting `docker-compose`.

Examples of migration files:
- `20251204120000_create_questions.sql` - creating the question table
- `20251204121000_create_answers.sql` - creating the answer table and foreign key with `ON DELETE CASCADE`

---

## API
## Models
### Question
```json
{
    "id": 1,
    "text": "How to use this API?",
    "created_at": "2025-12-04T12:00:00Z",
    "answers": [
        {
            "id": 1,
            "question_id": 1,
            "user_id": "user777",
            "text": "Just send requset",
            "created_at": "2025-12-04T12:05:00Z"
        }
    ]
}
```
### Answer
```json
{
    "id": 1,
    "question_id": 1,
    "user_id": "user777",
    "text": "Just send requset",
    "created_at": "2025-12-04T12:05:00Z"
}
```

---

## Questions endpoints 

## GET `/questions/`
Get a list of all qustions

Request:
```bash
curl -X GET http://localhost:8080/questions/
```
Response:
```json
[
    {
        "id": 1,
        "text": "First question",
        "created_at": "2025-12-04T12:00:00Z",
    },
    {
        "id": 2,
        "text": "Second question",
        "created_at": "2025-12-04T12:10:00Z",
    }
]
```

## POST `/questions/`
Create a new question

Request body:
```json
{
    "text": "New question"
}
```
Request:
```json
curl -X POST http://localhost:8080/questions/ \
    -H "Content-Type: application/json" \
    -d '{"text": "New question"}'
```
Response (`201 Created`):
```json
{
    "id": 3,
    "text": "New question",
    "created_at": "2025-12-04T13:00:00Z",
}
```

## GET `/questions/{id}`
Get a question and all the answers to it

Request:
```bash
curl -X GET http://localhost:8080/questions/1
```
Response (`200 OK`):
```json
{
    "id": 1,
    "text": "How to use this API?",
    "created_at": "2025-12-04T12:00:00Z",
    "answers": [
        {
            "id": 1,
            "question_id": 1,
            "user_id": "user777",
            "text": "Just send requset",
            "created_at": "2025-12-04T12:05:00Z"
        }
    ]
}
```

## DELETE `/questions/{id}`
Delete a question and all its answers

Request:
```bash
curl -X DELETE http://localhost:8080/questions/1
```
Response: `204 No Content`
If the question is not fount: `404 Not Found`

---

## Answers endpoints
## POST `/questions/{id}/answers/`
Add an answer to the question
You cant create an answer to a non-existent question (`404`)

Request body:
```json
{
    "user_id": "user777",
    "text": "My answer"
}
```
Request:
```bash
curl -X POST http://localhost:8080/questions/1/answers/ \
    -H "Content-Type: application/json" \
    -d '{"user_id": "user777", "text": "My answer"}'
```
Response (`201 Created`):
```json
{
    "id": 9,
    "question_id": 1,
    "user_id": "user777",
    "text": "My answer",
    "created_at": "2025-12-04T13:01:00Z",
}
```

## GET `/answers/{id}`
Get an answer by ID

Request:
```bash
curl -X GET http://localhost:8080/answers/10
```
Response (`200 OK`):
```json
{
    "id": 9,
    "question_id": 1,
    "user_id": "user777",
    "text": "My answer",
    "created_at": "2025-12-04T13:01:00Z",
}
```
If the answer is not found: `404 Not Found`

## DELETE `/answers/{id}`
Delete an answer by ID

Request:
```bash
curl -X DELETE http://localhost:8080/answers/10
```
Response: `204 No Content`
If the answer is not fount: `404 Not Found`

---

## Testing
Test POST request to create question and GET created question
Run:
```bash
go test ./internal/http/
```
