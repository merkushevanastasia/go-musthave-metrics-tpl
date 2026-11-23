```shell
migrate create -ext sql -dir ./migrations -seq create_metrics_table 
```

```shell
migrate -database "postgres://postgres_app_user:postgres_user_password@localhost:5432/postgres_app_db?sslmode=disable" -path ./migrations up 

```