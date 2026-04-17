# luhn-service

A minimal Go microservice that transforms an arbitrary number into one that passes the **Luhn algorithm** by appending the correct check digit.

---

## What is the Luhn Algorithm?

The [Luhn algorithm](https://en.wikipedia.org/wiki/Luhn_algorithm) (also known as the "modulus 10" algorithm) is a simple checksum formula used to validate credit card numbers, IMEI numbers, and other identification numbers.

Given a payload number, the algorithm computes a single **check digit** that is appended to the right of the payload. The resulting full number satisfies:

```
sum of all digits (with every second digit from the right doubled) ≡ 0 (mod 10)
```

This service accepts a raw number (no check digit), appends the correct check digit, and returns the Luhn-valid result.

---

## Endpoints

### `POST /luhn/transform`

Appends the Luhn check digit to the provided number.

**Request body (JSON):**
```json
{ "number": "7992739871" }
```

**Also accepts `application/x-www-form-urlencoded`:**
```
number=7992739871
```

**Response:**
```json
{
  "input":   "7992739871",
  "result":  "79927398713",
  "valid":   true,
  "message": "check digit appended — number now passes Luhn validation"
}
```

---

### `POST /luhn/validate`

Checks whether a full number (payload + existing check digit) passes the Luhn algorithm.

**Request body (JSON):**
```json
{ "number": "79927398713" }
```

**Response:**
```json
{
  "input":   "79927398713",
  "result":  "79927398713",
  "valid":   true,
  "message": "number passes Luhn check"
}
```

---

### `GET /health`

Health check.

```json
{ "status": "ok" }
```

---

## Quick start

### Run with Go

```bash
git clone https://github.com/ahmadbass3l/luhn-service.git
cd luhn-service
go run .
```

The service listens on `:8080` by default. Set `PORT` to override.

### Run with Docker

```bash
docker build -t luhn-service .
docker run -p 8080:8080 luhn-service
```

### Example call

```bash
curl -s -X POST http://localhost:8080/luhn/transform \
  -H "Content-Type: application/json" \
  -d '{"number": "7992739871"}' | jq
```

```json
{
  "input":   "7992739871",
  "result":  "79927398713",
  "valid":   true,
  "message": "check digit appended — number now passes Luhn validation"
}
```

---

## Running tests

```bash
go test -v -race ./...
```

All tests are in `main_test.go` and cover:

- Luhn checksum calculation (unit)
- `MakeLuhnValid` with known Wikipedia example (`7992739871` → `79927398713`)
- `IsLuhnValid` for valid and invalid numbers
- Separator stripping (spaces and hyphens)
- HTTP handlers: JSON body, form body, method guard, missing field

---

## Algorithm detail

```
Payload:  7  9  9  2  7  3  9  8  7  1
Position: 10 9  8  7  6  5  4  3  2  1   (1-indexed from right of final number)
Double?:  ✓     ✓     ✓     ✓     ✓
Doubled:  14 9  18 2  14 3  18 8  14 1
After -9: 5  9  9  2  5  3  9  8  5  1
Sum = 5+9+9+2+5+3+9+8+5+1 = 56 + 1 = 57... → (10 - (sum % 10)) % 10 = 3
Result:   79927398713  ✓
```

---

## Project structure

```
luhn-service/
├── main.go           # HTTP server + Luhn logic
├── main_test.go      # Unit and HTTP handler tests
├── go.mod            # Module definition
├── Dockerfile        # Multi-stage build (scratch image, ~5 MB)
└── .github/
    └── workflows/
        └── ci.yml    # GitHub Actions: test + build on every push
```

---

## Configuration

| Variable | Default | Description         |
|----------|---------|---------------------|
| `PORT`   | `8080`  | TCP port to listen on |

---

## License

MIT
