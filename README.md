# vk
 
### Run locally

```shell
docker-compose up -d postgres  
time sleep 10
docker-compose up filmlibrary
```

###  Tests
```shell
go test ./tests -coverprofile=coverage.out -coverpkg=./...
```

- Swagger available [here](./docs/swagger.yaml).
