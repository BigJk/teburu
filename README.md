# てぶる ー teburu

Self-hosted Google Sheets as an API. It's like a self-hosted version of [Sheety](https://sheety.co/) and [Sheet2API](https://sheet2api.com/).

## Idea

The basic idea is that if you had a google sheet table like that:

| name | description | price |
| ---- | ----------- | ----- |
| Apple | A fruit | 1.99 |
| Banana | A fruit | 2.99 |

You could access it as a JSON API like that:

```json
[
  {
    "name": "Apple",
    "description": "A fruit",
    "price": 1.99
  },
  {
    "name": "Banana",
    "description": "A fruit",
    "price": 2.99
  }
]
```

That way you can use Google Sheets as a CMS for your website or as a database for your app. This is especially useful if you have non-technical people who need to edit the data, or you need a quick prototyping solution without having to set up a management ui.

## Pre-requisites

### 1. Set up a Google Cloud project

You will need to set up a Google Cloud project to use teburu. This is because teburu uses the Google Sheets API to read and write to your spreadsheet. The free quota (300 requests per minute) should be enough for most use cases if you factor in caching. You can follow the steps below to set up a Google Cloud project:

- Go to the Google Cloud Console ([console.cloud.google.com](https://console.cloud.google.com)) and create a new project.
- Enable the Google Sheets API for your project by searching for "Google Sheets API" in the API Library and enabling it.
- Create API credentials by navigating to the "Credentials" section and clicking on "Create credentials" -> "Service account key". Select the appropriate service account and choose JSON as the key type. Download the JSON file that contains your credentials.
- Also save the email address (``*.gserviceaccount.com``) of the service account that you created. You will need this later.

### 2. Set up your spreadsheet

You will need to set up your spreadsheet to use teburu.

- Create a new spreadsheet in Google Sheets.
- Share the spreadsheet with the service account email address that you saved earlier. Give it edit access.
- In the first row of the spreadsheet, add your column names. These will be the keys in your JSON response.
- Now you can start adding data to your spreadsheet. You can add as many rows as you want. You can also add multiple sheets to your spreadsheet. Each sheet will be a separate endpoint in teburu.

## Endpoints

### GET /api/sheet/:id/:sheet

Returns the contents of the sheet as a JSON array.

#### Parameters

- ``id``: The ID of your spreadsheet. This is the long string of characters in the URL of your spreadsheet.
- ``sheet``: The name of the sheet in your spreadsheet.

### GET /api/sheet/:id/:sheet/:eid

Returns a single row from the sheet as a JSON object.

#### Parameters

- ``id``: The ID of your spreadsheet. This is the long string of characters in the URL of your spreadsheet.
- ``sheet``: The name of the sheet in your spreadsheet.
- ``eid``: The ID of the row to return. This is the number in the first column of your spreadsheet.

### Query Parameter

Additional query parameters can be added to the URL to modify the JSON response.

- ``case``: The case of the keys in the JSON response. Can be ``camel``, ``snake``, ``kebab``, ``plain`` and ``screaming_snake``. Defaults to ``camel``.
- ``columns``: The columns to return in the JSON response in comma seperated form like ``name,description``. Defaults to all columns.
- ``format``: The format of the JSON response. Can be ``simple``, ``dynamic`` and ``complex``. Defaults to ``simple``.
  - ``simple``: Returns row fields as simple ``"key": "value"`` pairs.
  - ``complex``: Returns the row fields as ``"key": { "value": "...", "hyperlink": "http://..." }`` where hyperlink contains the hyperlink if the cell contains one.
  - ``dynamic``: Returns the row fields as ``complex`` if the cell contains a hyperlink, otherwise as ``simple``.

## Install

Install teburu with one of the following methods and then run it with ``teburu``.

### Via Go

```go install github.com/BigJk/teburu/cmd/teburu@latest```

### Docker

You can also run teburu with docker. You can use the following command to run teburu with docker. Use the environment variables or the config.yaml file to configure teburu:

```bash 
docker run -p 8753:8753 \\
-v /location/sheet_creds.json:/creds.json \\
-e TEBURU_CACHE=1 \\
-e TEBURU_CACHE_TTL=5m0s \\
ghcr.io/bigjk/teburu:latest 
```

<details>
<summary>docker-compose.yml example</summary>

```yaml
version: "3.8"

services:
  teburu:
    image: ghcr.io/bigjk/teburu:latest
    ports:
      - 8753:8753
    volumes:
      - /location/sheet_creds.json:/creds.json
    environment:
      - TEBURU_CACHE=1
      - TEBURU_CACHE_TTL=5m0s
```

</details>

## config.yaml

After starting teburu for the first time, a config.yaml file will be created in the current directory. You can use this file to configure teburu.

```yaml
bind: :8753 # The address to bind the webserver to
cache: false # Whether to cache the requests
cache_ttl: 5m0s # The time to cache the requests for
cors: true # Whether to enable CORS
credentials_file: ./creds.json # The path to the credentials file from Google Cloud
rate_limit: 5 # The number of requests per minute to allow, 0 for no limit
request_timeout: 10s # The timeout for requests to the webserver, 0 for no timeout
gzip: true # Whether to enable gzip compression
```

### Environment Variables

Alternatively you can also use environment variables to configure teburu. The environment variables are prefixed with ``TEBURU_`` and are all uppercase. For example ``TEBURU_BIND`` or ``TEBURU_CACHE``.

## TODO

- [ ] Updating rows
- [ ] Deleting rows
- [ ] Adding rows
- [ ] Filtering rows
- [x] Docker
