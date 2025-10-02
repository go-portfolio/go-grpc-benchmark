# Сгенерировать Go код из proto
```bash
$ protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/benchmark.proto
```

# Запустить сервер
go run cmd/server/main.go

# Запустить клиент бенчмарка
go run cmd/client/main.go
