# Pulse Backend (minimal)

This is a minimal Node.js backend used for development. It provides a `/signup` endpoint and stores users in a local JSON file.

Quick start:

```bash
cd Backend
npm install
npm start
```

Endpoints:
- POST `/signup` - body: `{ "username": "u", "email": "e", "password": "p" }` -> 201 created
- GET `/users` - list of users (development convenience)

Data store: `persons.json` in the same folder.

Note: This is a simple dev server and does not implement production best practices.
