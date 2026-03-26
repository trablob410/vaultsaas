# Phase 10: API Documentation (OpenAPI)

**Priority:** P2 | **Effort:** 4h | **Status:** pending

Create OpenAPI/Swagger docs for all public API endpoints.

## Scope

Document all endpoints in:
- `GET/POST /secrets`
- `GET/PUT/DELETE /secrets/{id}`
- `POST /access-requests`
- `GET /access-requests`
- `POST /access-requests/{id}/approve`
- `POST /access-requests/{id}/reject`
- `GET /credentials/{id}`
- `POST /credentials/{id}/revoke`
- `GET /audit/logs`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- CLI-specific endpoints (if any)

## Documentation Format

Create `docs/api-reference.md` (OpenAPI 3.0 spec in YAML or JSON).

Include for each endpoint:

- **Summary:** One-line description
- **Parameters:** Query, path, body (with types)
- **Response:** Success (200/201) and error codes (400, 401, 403, 404, 429, 500)
- **Authentication:** Bearer token required?
- **Rate Limit:** Subject to 60 rpm limit?
- **Example Request/Response**
- **Notes:** Special behavior, edge cases

Example format:

```markdown
## Create Secret

**POST** `/api/v1/secrets`

**Authentication:** Required (Bearer token)

**Body:**
```json
{
  "name": "string",
  "value": "string",
  "project_id": "uuid",
  "encrypted_blob": "base64",
  "encrypted_dek": "base64"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "name": "string",
  "project_id": "uuid",
  "created_at": "RFC3339",
  "owner_user_id": "uuid"
}
```

**Errors:**
- `400 Bad Request` — Missing required fields
- `401 Unauthorized` — Invalid token
- `403 Forbidden` — User not member of project
- `409 Conflict` — Secret name already exists
```

## Tools (Optional)

- **Swagger UI:** Host OpenAPI spec with interactive explorer
- **ReDoc:** Alternative, cleaner UI for API docs
- **Postman:** Generate from OpenAPI spec for easy testing

## Checklist

- [ ] All 15+ endpoints documented
- [ ] All request/response fields documented
- [ ] Error codes documented (400, 401, 403, 404, 429, 500)
- [ ] Authentication requirements clear
- [ ] Rate limits documented
- [ ] Example requests/responses provided
- [ ] CLI special endpoints (if any) documented
- [ ] Docs reviewed for accuracy
- [ ] Docs version-controlled in `docs/`

## Publication

1. Save OpenAPI spec to `docs/openapi.yaml`
2. Option A: Host Swagger UI at `/api/docs`
3. Option B: Link to spec in README
4. Share with early beta users

## Notes

- Keep docs in sync with code (review during code review)
- Version docs when API changes (v1.0, v1.1, etc.)
- Include deprecation notices for old endpoints
