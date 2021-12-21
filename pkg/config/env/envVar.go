package env

import "time"

const (
	SymmetricKey  = "12345678901234567890123456789012"
	TokenDuration = 15 * time.Minute
)

type Config struct {
	DbDriver          string
	DbSource          string
	ServerAddress     string
	TokenSymmetricKey string
	TokenDuration     time.Duration
}

func NewConfig(symmetricKey string, tokenDuration time.Duration) Config {
	return Config{
		DbDriver:          "postgres",
		DbSource:          "postgresql://root:root@localhost:5432/recipes?sslmode=disable",
		ServerAddress:     "0.0.0.0:8080",
		TokenSymmetricKey: symmetricKey,
		TokenDuration:     tokenDuration,
	}
}
