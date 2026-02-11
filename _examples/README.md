# JSON:API examples

Two examples demonstrate the library with a flat-file database, validation, links, and relationships.

## Server

The server exposes JSON:API resources **stores** and **pets** with full CRUD. Data is stored in `_examples/server/data/` as JSON files.

- **Stores**: `name` (required, 1–100 chars), `address` (optional, max 200 chars). Relationship: `pets` (to-many).
- **Pets**: `name` (required, 1–80 chars), `species` (required: `dog`, `cat`, or `bird`). Relationship: `store` (to-one).

Run from the repo root:

```bash
go run ./_examples/server
```

Listens on **http://localhost:8080**. Endpoints:

- `GET/POST /stores`, `GET/PATCH/DELETE /stores/:id`
- `GET/POST /pets`, `GET/PATCH/DELETE /pets/:id`

Responses include `links.self` and relationship links (`self` and `related`). Validation errors return `422` with JSON:API `errors[]` and `source.pointer` (e.g. `/data/attributes/name`).

## Client

The client is a static UI (HTML + JS) that consumes the server API and displays validation errors next to form fields.

Run from the repo root (with the server already running):

```bash
go run ./_examples/client
```

Open **http://localhost:8081**. Use the forms to create/edit stores and pets; invalid input shows field-level errors from the server’s validation response.

## Quick test

1. Start server: `go run ./_examples/server`
2. Start client: `go run ./_examples/client`
3. Open http://localhost:8081
4. Create a store, then a pet linked to that store
5. Submit invalid data (e.g. empty name, or species "fish") to see validation errors in the UI
