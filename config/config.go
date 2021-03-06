package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	cfg *Config
	// Version application version injected in build
	Version string // nolint
)

// Config represents all application settings
type Config struct {
	ApplicationName string    `yaml:"application_name" env:"IAM_APPLICATION_NAME" env-default:"lab/iam"`
	BindAddress     string    `yaml:"bind_address" env:"IAM_BIND_ADDRESS" env-default:":9000"`
	Production      bool      `yaml:"production" env:"IAM_PRODUCTION" env-default:"false"`
	LogLevel        int       `yaml:"log_level" env:"IAM_LOG_LEVEL" env-default:"0"`
	LogPath         string    `yaml:"log_path" env:"IAM_LOG_PATH" env-default:"/var/log/lab/iam"`
	LogSTDOUT       bool      `yaml:"log_stdout" env:"IAM_LOG_STDOUT" env-default:"false"`
	Database        database  `yaml:"database"`
	JWT             JWTConfig `yaml:"jwt"`
	Version         string
	PublicKeys      map[string]*rsa.PublicKey
}

// JWTConfig holds all JWT-related settings
type JWTConfig struct {
	Issuer   string `yaml:"issuer" env:"IAM_JWT_ISSUER" env-default:"lab/iam"`
	Audience string `yaml:"audience" env:"IAM_JWT_AUDIENCE"`
	Keys     []struct {
		ID          string `yaml:"id"`
		PrivatePath string `yaml:"private"`
		PublicKey   *rsa.PublicKey
		PrivateKey  *rsa.PrivateKey
	} `yaml:"keys"`
}

type database struct {
	MinimunMigration string `yaml:"minimum_migration" env:"IAM_DATABASE_MINIMUM_MIGRATION"`
	RWDatabase       string `yaml:"rw_database_uri" env:"IAM_RW_DATABASE"`
	RODatabase       string `yaml:"ro_database_uri" env:"IAM_RO_DATABASE"`
	MaxConnLifetime  int    `yaml:"max_conn_lifetime" env:"IAM_DATABASE_MAX_CONN_LIFETIME"`
	MaxOpenConn      int    `yaml:"max_open_conn" env:"IAM_DATABASE_MAX_OPEN_CONN"`
	MaxIdleConn      int    `yaml:"max_idle_conn" env:"IAM_DATABASE_MAX_IDLE_CONN"`
}

func init() {
	var (
		cfgPath = "config.yaml"
		err     error
	)

	if val := os.Getenv("IAM_CONFIG_FILE"); val != "" {
		cfgPath = val
	}

	cfg = new(Config)

	if err = cleanenv.ReadConfig(cfgPath, cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.Database.MinimunMigration == "" {
		log.Fatal("unable to start without minimum migration set")
	}

	if cfg.Database.RODatabase == "" {
		cfg.Database.RODatabase = cfg.Database.RWDatabase
	}

	if Version == "" {
		Version = time.Now().Format("200601021504") + "-development"
	}

	if cfg.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg.PublicKeys = make(map[string]*rsa.PublicKey)

	for i := range cfg.JWT.Keys {
		var (
			rawKey []byte
			block  *pem.Block
		)

		if rawKey, err = ioutil.ReadFile(cfg.JWT.Keys[i].PrivatePath); err != nil {
			log.Fatal(err)
		}

		block, _ = pem.Decode(rawKey)

		if block.Type != "RSA PRIVATE KEY" {
			log.Fatal("Invalid private key type")
		}

		if cfg.JWT.Keys[i].PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
			log.Fatal(err)
		}

		cfg.JWT.Keys[i].PublicKey = &cfg.JWT.Keys[i].PrivateKey.PublicKey

		cfg.PublicKeys[cfg.JWT.Keys[i].ID] = cfg.JWT.Keys[i].PublicKey
	}

	cfg.Version = Version
}

// GetConfig returns a copy of running settings
func GetConfig() Config {
	if cfg == nil {
		log.Fatal("config not loaded")
	}

	return *cfg
}
