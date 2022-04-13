# Geolocation service

This is a simple geolocation service for the FindHotel Coding Challenge. [Requirements](task/README.md)

## Development

`local.env` contains env variables for configuring the service

Makefile contains most needed aliases

* `make docker.all.start` - start database container and API server
* `make import.run` - run import service
* `make test` - run unit tests

After you start database, API and finish import, you can start using the service. Sample request using `curl`:

```bash
curl --location --request GET 'http://localhost:8080/geolocation/160.103.7.140'
```

## Import details

* Records are validated before inserting.
    * IP adress must be valid
    * Country code must be 2 letters long
    * Country, City must be not empty
    * Latitude: float [-90, 90]
    * Longitude: float [-180, 180]
    * Mystery value: any
* Deduplication process: first ip geoinfo to be inserted in DB is kept, all other records are discarded.
* Statictics numbers are returned after import:
    * records inserted
    * duplicated records in DB
    * duplicated records in CSV
    * Unparsed records
    * Time elapsed

## API

OpenAPI schema: [api.yaml](./api.yaml)
