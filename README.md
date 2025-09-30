# IP Query Documentation

A Go-based service for querying geographic location information by IP address.

## Database Requirement

The project does not include database files. You need to download the required database from [lite.ip2location.com](https://lite.ip2location.com/database-download "IP2LOCATION-LITE-DB3.BIN download").

## Features

- Simple REST API endpoint
- Returns detailed location information, including country, region, and city
- Easy to set up and run

## Installation for Development

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd ipquery
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the service:
   ```bash
   go run .
   ```

The service will start running on `http://localhost:8080`.

## Usage

There are two ways to use the service:

### API Service Mode

Query location information by making a GET request to the `/query` endpoint.  
- If no `api_key` is specified, the API request only requires the IP address.  
- If an `api_key` environment variable is set, the API request must include a `sign` (signature) and `time` (timestamp) in addition to the IP address.  
- Alternatively, set the environment variable `IPQUERY_API_KEY=xxxx`.

Start the service:

```bash
ipquery -port :80 -db_path ./IP2LOCATION-LITE-DB3.BIN
```

Or with an API key:

```bash
ipquery -api_key=xxx -port :80 -db_path ./IP2LOCATION-LITE-DB3.BIN
```

Make API requests:

```bash
curl "http://localhost:8080/query?ip=202.106.0.20"
curl "http://localhost:8080/query?ip=202.106.0.20&sign=xxx&time=2020303434"
```

### Command-Line Mode

```bash
ipquery -query 202.106.0.20
```

### Response Format

The service returns location data in JSON format:

```json
[
  {
    "ip": "202.106.0.20",
    "country": "China",
    "country_code": "CN",
    "region": "Beijing",
    "city": "Beijing"
  }
]
```

### Response Fields

- `ip`: The queried IP address
- `country`: Country name
- `country_code`: Two-letter country code (ISO 3166-1 alpha-2)
- `region`: Region or province name
- `city`: City name

## API Endpoint

**GET** `/query?ip=<ip_address>`

- **Parameters**:
  - `ip` (required): The IP address to query (IPv4 format). Multiple IPs can be separated by commas, returning an array of location objects.

- **Response**: A JSON object containing location information. For multiple IPs, an array of location objects is returned.

**Examples**:

- Single IP request:
  ```
  /query?ip=34.32.23.12
  ```
  ```json
  {"ip":"34.32.23.12","country":"Germany","country_code":"DE","region":"Berlin","city":"Berlin"}
  ```

- Multiple IP request:
  ```
  /query?ip=34.32.23.12,34.32.23.13
  ```
  ```json
  [
    {"ip":"34.32.23.12","country":"Germany","country_code":"DE","region":"Berlin","city":"Berlin"},
    {"ip":"34.32.23.13","country":"Germany","country_code":"DE","region":"Berlin","city":"Berlin"}
  ]
  ```

## Development

To build the application:

```bash
go build -o ipquery
```

## Using Docker or Podman

Run the service with a container:

```bash
podman run --rm -it -p 8080:8080 -v ./IP2LOCATION-LITE-DB3.BIN:/app/IP2LOCATION-LITE-DB3.BIN javyliu/ipquery:v1.0.1
```

To specify an `IPQUERY_API_KEY` environment variable:

```bash
podman run --rm -it -p 8080:8080 -v ./IP2LOCATION-LITE-DB3.BIN:/app/IP2LOCATION-LITE-DB3.BIN -e IPQUERY_API_KEY=xxxx javyliu/ipquery:v1.0.1
```

## TODO

- Add support for multi-language display

## License

IpQuery is released under the MIT License.