wallet_db:
	docker run --name wallet_db -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=wallet -d postgres

wallet_db_migrateup:
	goose -dir "db/wallet/migrate" postgres "user=postgres password=postgres host=localhost dbname=wallet sslmode=disable" up

wallet_db_migratedown:
	goose -dir "db/wallet/migrate" postgres "user=postgres password=postgres host=localhost dbname=wallet sslmode=disable" down

wallet_db_sqlc_generate:
	cd db/wallet && sqlc generate

wallet_service:
	go run cmd/api/wallet_service/*.go

scanner_service:
	go run cmd/api/scanner_service/*.go

amqp:
	docker run --name amqp -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=guest -e RABBITMQ_DEFAULT_PASS=guest -d rabbitmq:3-management

.PHONY: wallet_db wallet_db_migrateup wallet_db_migratedown wallet_db_sqlc_generate wallet_service scanner_service amqp
