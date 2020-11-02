module lab/iam

go 1.15

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/google/uuid v1.1.2
	github.com/ilyakaznacheev/cleanenv v1.2.5
	github.com/jackc/pgconn v1.7.1
	github.com/jackc/pgx/v4 v4.9.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/washingt0/oops v0.0.1
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/text v0.3.4 // indirect
)

replace github.com/washingt0/oops => ../../oops
