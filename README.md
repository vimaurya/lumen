
# Lumen 

Lumen is a high-performance **API Gateway** and **Layer 7 Load Balancer** written in Go. It serves as a secure entry point for microservices, providing automated traffic distribution, rate limiting, and deep request analytics with an integrated dashboard.

## System Architecture

Lumen sits between your users and your backend services. It processes every request through a pipeline of internal modules before forwarding it to a healthy backend node.

1.  **Security Layer:** Implements IP-based rate limiting to prevent DDoS and brute-force attacks.
    
2.  **Analytics Engine:** Asynchronously collects metadata (GeoIP, User-Agent, Latency) for every request.
    
3.  **Load Balancer:** Distributes traffic across a pool of backend servers using a Round-Robin strategy.
    
4.  **Reverse Proxy:** Forwards the sanitized request to the backend with a secure `X-Lumen-Secret` header to ensure requests cannot bypass the gateway.
    

----------

## Getting Started

### Installation

You can build Lumen for multiple platforms using the included build script:

```bash
chmod +x build.sh
./build.sh

```

This generates binaries in the `/dist` folder for Linux, Windows, and macOS (Intel/M1).

### Initial Configuration

Lumen includes an interactive CLI to help you generate your initial `config.yaml`.
```bash
./lumen init

```

During this process, you will be prompted to:

-   Enter your target backend URLs (e.g., `http://localhost:8081, http://localhost:8082`).
    
-   Set a **Secret Token** that Lumen will use to authenticate itself to your backend services.
    

----------

## Features & Usage

### 1. Layer 7 Load Balancing

Lumen automatically balances traffic across the targets defined in your config.

-   **Strategy:** Round-Robin (Atomic-based index switching).
    
-   **Path Preservation:** You can configure whether the gateway should keep the original URL path or strip the prefix before forwarding.
    

### 2. Security & Rate Limiting

The gateway includes a built-in protective layer to keep your services stable:

-   **Rate Limiting:** Defaults to a generous threshold (15,000 requests) with a "cool-down" period that grants a bonus limit to recovered IPs.
    
-   **Admin Protection:** Access to the analytics dashboard is guarded by Basic Auth.
    
-   **Asset Exclusion:** Common static files (e.g., `.css`, `.js`, `.jpg`) can be ignored by the analytics engine to save resources.
    

### 3. Real-time Analytics Dashboard

Lumen provides a full-featured UI for monitoring your infrastructure health and visitor behavior.

-   **System Health:** View total views, unique IP visitors, error counts, and average latency (ms).
    
-   **Visitor Engagement:** Track active sessions, average session time, and bounce rates.
    
-   **Geolocation:** Real-time mapping of visitor countries using the GeoLite2 database.
    

----------

## Configuration Reference (`config.yaml`)

**Field**

**Description**

`server.port`

The port Lumen listens on (default: 8080).

`server.admin_path`

The URL path for your analytics dashboard.

`security.lumen_secret`

The token sent in the `X-Lumen-Secret` header to backends.

`proxy.targets`

List of backend URLs to balance traffic across.

`proxy.preserve_path`

If true, the full path is sent to the backend.

----------

## Testing the Setup

You can use the included `testapi.go` to simulate a backend service and verify that Lumen is correctly forwarding headers. Additionally, the `simulation/traffic.py` script allows you to stress-test the gateway with thousands of concurrent requests to see the rate limiter and analytics in action.
