# JQ-API

This web service get a json from an url then parse it with jq (check <https://github.com/stedolan/jq>).

## Installation

Build the service with this command:

```sh
docker build -t flyvolcano/jq-api .
```

and run the container:
```sh
docker run -p 8080:8080 flyvolcano/jq-api --restart unless-stopped
```

## Usage

Api params, can be url encoded:

- `url`: the url of your json
- `jq`: the jq query

The headers, body and request type used to request the api will be used to request the `url`.

Example :

```bash
curl 'http://localhost:8080/?jq=%22[.[]|{species:%20.[\%22Species%20Name\%22]}]%22&url=https://www.fishwatch.gov/api/species'
```

## Reserved characters

The `+` character which is used in jq is escaped by the query parameter. So as an alternative
the `±` character is used which internally gets converted to `+` before passing into jq.

## Improvements
1. Handles `+` via `±` character
2. Handles gziped responses
3. Migrated from gopkg to go modules