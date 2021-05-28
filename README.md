# itsm-commenting-service
ITSM commenting service

`make test` runs unit tests

`make e2e-test` runs end-to-end integration tests (requires docker-compose installed)

`make test-all` runs all tests

`make run` starts application for local use/testing

`make docs` starts API documentation server on default port 3001;
you can specify different port: `make docs PORT=3002`

`make swagger` regenerates swagger.yaml file from source code (usually no need to use unless API changes)
