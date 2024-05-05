wallet_db:
	docker run --name wallet_db -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=wallet -d postgres

wallet_db_migrateup:
	goose -dir "db/wallet/migrate" postgres "user=postgres password=postgres host=localhost dbname=wallet sslmode=disable" up

wallet_db_migratedown:
	goose -dir "db/wallet/migrate" postgres "user=postgres password=postgres host=localhost dbname=wallet sslmode=disable" down

wallet_db_sqlc_generate:
	cd db/wallet && sqlc generate

.PHONY: wallet_db wallet_db_migrateup wallet_db_migratedown wallet_db_sqlc_generate
