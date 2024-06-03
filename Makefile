include .env
.Phony: run/api
run/api:
	go run ./cmd/api -db-dsn=${DSN}

.Phony: db/psql
db/psql:
	psql ${DSN}

.Phony: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext .sql -dir migrations ${name}

.Phony: db/migrations/up
db/migrations/up:
	@echo 'Running Up migrations'
	migrate -path ./migrations -database ${DSN} up

.Phony: audit
audit:
	@echo 'Running audit'
	go mod tidy
	go mod verify
	@echo 'Formatting code'
	go fmt ./...
	@echo 'Vetting code'
	go vet ./...
	@echo 'Running Tests'
	go test -race -vet=off ./...
	##staticcheck ./...

.Phony: vendor
vendor:
	@echo 'Running Vendor'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies'
	go mod vendor

.Phony: build/api
build/api:
	@echo 'Building binaries'
	go build -o=./bin/api.exe ./cmd/api 
	set GOOS=linux
	set GOARCH=amd64
	go build -o ./bin/linux_amd64/api ./cmd/api
 

.Phony: run/bin
run/bin:
	.\bin\api.exe -db-dsn=${DSN}

.Phony: production/migrations/up
production/migrations/up:
	scp -i ./movie-backend-key.pem -r migrations greenlight@13.53.193.190:/home/greenlight/


.Phony: production/configure/api.service
production/configure/api.service:
	scp -i ./movie-backend-key.pem -r remote/setup/api.service greenlight@13.53.193.190:~
	
