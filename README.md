## GDELT Real-time Event Analysis API

---

## Features

### Real-time Data Integration
- **Automatic updates every 15 minutes** from the GDELT feed  
- **Historical data retention** in a SQLite database  
- **Daily automatic reset** of data older than 24 hours

### Geographic Intelligence
- **Country boundary mapping** using GeoJSON  
- **Automatic coordinate resolution** for event locations  
- **Multi-polygon spatial calculations** for complex boundaries

### Sentiment Analysis
- **Goldstein Scale** metric integration  
- **Average Tone** calculations per event  
- **News volume tracking** (mentions, sources, articles)

---

## Technology Stack

- **Go 1.21+**
- **SQLite 3**
- **GeoJSON** for geographic data
- **REST API** architecture

---

## Installation

### Prerequisites

```bash
# Install Go SQLite driver
go get github.com/mattn/go-sqlite3

# Install project dependencies
go mod tidy
```

### Setup

```bash
git clone https://github.com/yourusername/gdelt-api.git
cd gdelt-api
```

---

## Configuration

Create a file named `.env` in the project root with the following contents:

```ini
PORT=8080
GEOJSON_PATH=./config/countries.geo.json
DB_PATH=./data/events.db
```

---

## API Documentation

### Base URL

```
http://localhost:8080/gdelt/
```

### Endpoints

---

#### 1. Get Events

```
GET /events
```

**Parameters**:

| Parameter       | Type    | Description                               |
|-----------------|---------|-------------------------------------------|
| `country`       | string  | Filter by full country name              |
| `sourceCountry` | string  | Filter by the source actor's country code |
| `startDate`     | date    | Start date (YYYYMMDD)                     |
| `endDate`       | date    | End date (YYYYMMDD)                       |
| `minGoldstein`  | float   | Minimum Goldstein scale (-10 to 10)       |
| `maxGoldstein`  | float   | Maximum Goldstein scale                   |
| `minTone`       | float   | Minimum average tone                      |
| `maxTone`       | float   | Maximum average tone                      |
| `minLat`/`maxLat` | float | Latitude range for bounding box           |
| `minLng`/`maxLng` | float | Longitude range for bounding box          |
| `page`          | int     | Pagination index (default: 0)             |
| `limit`         | int     | Items per page (default: 50)              |
| `all`           | bool    | Return all results (ignores pagination)   |

**Response**:
```json
{
  "page": 0,
  "results": [
    {
      "GlobalEventID": "1223292259",
      "Date": "20240129",
      "SourceActor": {
        "Code": "CHN",
        "Name": "CHINA",
        "CountryCode": "CHN"
      },
      "TargetActor": {
        "Code": "",
        "Name": "",
        "CountryCode": ""
      },
      "EventCode": "051",
      "EventRootCode": "05",
      "GoldsteinScale": 3.4,
      "AvgTone": 6.31120783460283,
      "NumMentions": 1,
      "NumSources": 1,
      "NumArticles": 1,
      "SourceURL": "https://nowgrenada.com/2025/01/20th-anniversary-celebration-of-resumption-of-diplomatic-relations-coincides-with-chinese-new-year/",
      "Lat": 29.8782,
      "Lng": 121.549,
      "Country": "China"
    }
  ]
}
```

---

## Database Schema

Below is the recommended schema for the `events` table in `SQLite`:

```sql
CREATE TABLE events (
  id INTEGER PRIMARY KEY,
  global_event_id TEXT,
  event_date TEXT,
  source_actor TEXT,
  target_actor TEXT,
  lat REAL,
  lng REAL,
  country TEXT,
  goldstein_scale REAL,
  avg_tone REAL,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Automatic Maintenance

- **Data older than 24 hours** is automatically purged.  
- **Database vacuum** runs daily at **00:00 UTC**.  
- **Connection pooling** with a maximum of 100 connections.

---

## Query Examples

1. **Recent events in Ukraine**:
   ```bash
   curl "http://localhost:8080/gdelt/events?country=Ukraine"
   ```

2. **Today's events**:
   ```bash
   curl "http://localhost:8080/events"
    ```
3. **Events in China with a Goldstein scale > 5**:
   ```bash
   curl "http://localhost:8080/gdelt/events?country=China&minGoldstein=5"
   ```

**Note**: An active internet connection is required for GDELT data updates. Geographic resolution accuracy is dependent on the quality of the underlying GeoJSON data.
```