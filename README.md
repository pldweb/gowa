# GoWA - WhatsApp Gateway Multidevice

A lightweight WhatsApp gateway API built with Go, supporting WhatsApp's multidevice protocol. This gateway allows you to send and receive WhatsApp messages through a simple REST API.

## Features

- ✅ WhatsApp Multidevice support using [whatsmeow](https://github.com/tulir/whatsmeow)
- ✅ RESTful API for easy integration
- ✅ QR code authentication for device pairing
- ✅ Send text messages
- ✅ Connection status monitoring
- ✅ Session persistence
- ✅ Graceful shutdown

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

### From Source

```bash
git clone https://github.com/pldweb/gowa.git
cd gowa
go mod download
go build -o gowa
```

## Configuration

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `3000` |
| `SESSION_PATH` | Path to store session data | `./sessions` |
| `WEBHOOK_URL` | Webhook URL for incoming messages (optional) | - |

Example `.env` file:

```env
PORT=3000
SESSION_PATH=./sessions
WEBHOOK_URL=https://your-webhook-url.com/webhook
```

## Usage

### Starting the Server

```bash
./gowa
```

Or with environment variables:

```bash
PORT=8080 SESSION_PATH=/var/gowa/sessions ./gowa
```

### API Endpoints

#### 1. Health Check

Check if the service is running.

```bash
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "gowa"
}
```

#### 2. Get QR Code for Pairing

Get a QR code to pair a new device with WhatsApp.

```bash
GET /qr
```

**Response (when not connected):**
```json
{
  "qr_code": "2@...",
  "qr_image": "base64_encoded_png_image"
}
```

**Response (when already connected):**
```json
{
  "error": "Already connected"
}
```

**Note:** Scan the QR code with WhatsApp on your phone to pair the device.

#### 3. Check Connection Status

Check if the gateway is connected to WhatsApp.

```bash
GET /status
```

**Response:**
```json
{
  "connected": true,
  "jid": "1234567890@s.whatsapp.net"
}
```

#### 4. Send Message

Send a text message to a WhatsApp number.

```bash
POST /send
Content-Type: application/json

{
  "phone": "1234567890",
  "message": "Hello from GoWA!"
}
```

**Response:**
```json
{
  "status": "sent",
  "message_id": "MESSAGE_ID_HERE"
}
```

**Error Response:**
```json
{
  "error": "Not connected"
}
```

#### 5. Logout

Logout from WhatsApp and clear session data.

```bash
POST /logout
```

**Response:**
```json
{
  "status": "logged out"
}
```

## Example Usage with cURL

### 1. Check Health
```bash
curl http://localhost:3000/health
```

### 2. Get QR Code for Pairing
```bash
curl http://localhost:3000/qr
```

### 3. Check Status
```bash
curl http://localhost:3000/status
```

### 4. Send Message
```bash
curl -X POST http://localhost:3000/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "1234567890",
    "message": "Hello from GoWA!"
  }'
```

### 5. Logout
```bash
curl -X POST http://localhost:3000/logout
```

## Development

### Building

```bash
go build -o gowa
```

### Running Tests

```bash
go test ./...
```

### Code Structure

```
gowa/
├── main.go                 # Application entry point and HTTP handlers
├── internal/
│   ├── config/
│   │   └── config.go      # Configuration management
│   └── gateway/
│       └── gateway.go     # WhatsApp client wrapper
├── sessions/              # Session data storage (auto-created)
└── go.mod                 # Go module dependencies
```

## Security Considerations

- Keep your session data secure. The `sessions/` directory contains sensitive authentication data.
- Use HTTPS in production environments.
- Implement proper authentication/authorization for API endpoints in production.
- Never commit session files to version control (included in `.gitignore`).

## Troubleshooting

### Connection Issues

1. Make sure WhatsApp Web is not open in any browser
2. Try logging out and getting a new QR code
3. Check that session directory has proper permissions

### QR Code Not Generating

1. Ensure you're not already connected (check `/status`)
2. Try restarting the service
3. Delete the `sessions/` directory and try again

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- [whatsmeow](https://github.com/tulir/whatsmeow) - WhatsApp Web multidevice library for Go
- [Fiber](https://github.com/gofiber/fiber) - Express-inspired web framework for Go