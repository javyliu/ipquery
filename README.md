# IP Query

A Go-based service for querying geographic location information by IP address.

## Features

- Simple REST API endpoint
- Returns detailed location information including country, region, and city
- Easy to set up and run

## Installation for develop

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

Query location information by making a GET request to the `/query` endpoint:

```bash
curl "http://localhost:8080/query?ip=202.106.0.20"
```

### Response Format

The service returns location data in JSON format:

```json
[
  {
    "ip": "202.106.0.20",
    "country": "中国",
    "country_code": "CN",
    "region": "北京市",
    "city": "北京"
  }
]
```

### Response Fields

- `ip`: The queried IP address
- `country`: Country name (in Chinese)
- `country_code`: Two-letter country code (ISO 3166-1 alpha-2)
- `region`: Region or province name
- `city`: City name

## API Endpoint

**GET** `/query?ip=<ip_address>`

- **Parameters**: 
  - `ip` (required): The IP address to query (IPv4 format)

- **Response**: JSON array containing location objects

## Development

To build the application:

```bash
go build -o ipquery
```

Command line

```bash
ipquery -query 202.106.0.20
```

To start with a API

```bash
ipquery
```

## License

IpQuery is released under the MIT License.