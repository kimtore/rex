rex:
	go build -o bin/rex cmd/rex/*.go

sql:
	sqlc generate
