# Server Room Temperature

A lightweight monitoring system for server room temperature telemetry.

## Architecture

The system consists of three main components:

1. srt-sensor: Firmware written in TinyGo for the Raspberry Pi Pico (rp2040) that reads temperature and streams binary frames over USB CDC.
2. srt-consumer: Host utility that reads framed temperature measurements from the serial port (/dev/ttyACM0) and publishes JSON telemetry to a NATS broker.
3. srt-dashboard: History API and single-page application dashboard served with pre-rendered configuration injection and CORS support.

Shared infrastructure (NATS broker with JetStream, stream bootstrap hook, and the dashboard) is packaged in the Helm chart under chart/.

## Building with gob

This repository uses [gob](https://github.com/phillezi/gob) as the build recipe tool.

### Build Targets

Run build targets directly with go run gob.go:

- Build all binaries to bin/ (host binaries and sensor firmware):

  ```bash
  go run gob.go all
  ```

- Build only host binaries (srt-consumer and srt-dashboard):

  ```bash
  go run gob.go host
  ```

- Build only Raspberry Pi Pico sensor firmware (requires TinyGo):

  ```bash
  go run gob.go sensor
  ```

- Clean build artifacts:

  ```bash
  go run gob.go clean
  ```

## Usage

### Running srt-consumer

The consumer connects to the sensor serial port and forwards data to NATS.

```bash
./bin/srt-consumer --device /dev/ttyACM0 --nats-url nats://localhost:4222
```

Available flags and environment variables:

- --device, -d: Serial device path (default: /dev/ttyACM0, env: SERIAL_PORT, DEVICE)
- --nats-url: NATS server URL (default: nats://localhost:4222, env: NATS_URL)
- --nats-host: NATS server host (default: localhost:4222, env: NATS_HOST)
- --nats-user: NATS username (default: temperature-writer, env: NATS_USER)
- --nats-password: NATS password (env: NATS_PASSWORD)
- --topic, -t: NATS subject to publish to (default: serverroom.temperature.room1.sensor1, env: NATS_TOPIC)

### Running srt-dashboard

The dashboard serves both the frontend static UI and the /api/history endpoint querying JetStream.

```bash
./bin/srt-dashboard --http-addr :8080 --nats-host localhost:4222
```

Available flags and environment variables:

- --http-addr, -a: HTTP server listen address (default: :8080, env: HTTP_ADDR, PORT)
- --nats-host: NATS server host (default: localhost:4222, env: NATS_HOST)
- --nats-user: Backend NATS username (default: temperature-dashboard, env: NATS_USER)
- --nats-password: Backend NATS password (env: NATS_PASSWORD)
- --nats-ws-url: NATS WebSocket URL injected to browser client (default: ws://localhost:9222, env: NATS_WS_URL)
- --nats-reader-user: Reader NATS username for browser (default: temperature-reader, env: NATS_READER_USER)
- --nats-reader-password: Reader NATS password for browser (env: NATS_READER_PASSWORD)
- --stream, -s: JetStream stream name (default: SERVER_ROOM_TEMPERATURE, env: STREAM_NAME)
- --topic, -t: Default subject for browser client (default: serverroom.temperature.room1.sensor1, env: NATS_SUBJECT)

## Kubernetes Deployment (Helm)

The chart in chart/ deploys NATS with JetStream storage, a stream creation hook, and the dashboard.

```bash
helm install srt ./chart
```

Enable Ingress:

```bash
helm install srt ./chart \
  --set dashboard.ingress.enabled=true \
  --set dashboard.ingress.hosts[0].host=srt.example.com \
  --set nats.websocket.ingress.enabled=true \
  --set nats.websocket.ingress.hosts[0].host=nats-ws.example.com
```

Or enable Gateway API (HTTPRoute):

```bash
helm install srt ./chart \
  --set dashboard.gateway.enabled=true \
  --set dashboard.gateway.hostnames[0]=srt.example.com \
  --set nats.websocket.gateway.enabled=true \
  --set nats.websocket.gateway.hostnames[0]=nats-ws.example.com
```
