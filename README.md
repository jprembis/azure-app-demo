# Azure app demo

## Description

A JSON over HTTP server hosted at <https://kloda.azurewebsites.net> that exposes endpoints for writing to and reading from an append-only log.

## Usage

Append a record with a Base64-encoded value:

	$ curl -X POST https://kloda.azurewebsites.net -d '{"record":{"value":"SGVsbG8gQXp1cmUK"}}'

Lookup a record by index:

	$ curl -X GET https://kloda.azurewebsites.net -d '{"key":0}'

## Sources

Adapted from [[1](https://https://github.com/travisjeffery/proglog/LetsGo)] and [[2](https://github.com/edandersen/go-azure-appservice)].
