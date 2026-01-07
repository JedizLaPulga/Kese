# Tutorial Example - Kese Todo API

This directory contains the complete code from the Kese tutorial.

## What's Inside

- `main.go` - Main application with all route handlers
- `todo.go` - Todo model and in-memory store

## How to Run

```bash
# From the tutorial directory
go run main.go todo.go

# Or build and run
go build -o tutorial
./tutorial
```

Server will start on `http://localhost:8080`

## API Endpoints

- `GET /` - API information
- `GET /todos` - List all todos
- `GET /todos/:id` - Get a specific todo
- `POST /todos` - Create a new todo
- `PUT /todos/:id` - Update a todo
- `DELETE /todos/:id` - Delete a todo

## Example Requests

### Create a todo:
```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Kese Framework"}'
```

### Get all todos:
```bash
curl http://localhost:8080/todos
```

### Update a todo:
```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Master Kese Framework","completed":true}'
```

### Delete a todo:
```bash
curl -X DELETE http://localhost:8080/todos/1
```

## Features Demonstrated

- ✅ RESTful routing
- ✅ Middleware (Logger, Recovery, CORS, RequestID)
- ✅ JSON request/response handling
- ✅ URL parameters (`:id`)
- ✅ Query parameters (`?completed=true`)
- ✅ Error handling
- ✅ Thread-safe data store

See the full tutorial at [docs/TUTORIAL.md](../../docs/TUTORIAL.md)
