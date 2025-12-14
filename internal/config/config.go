package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Firebase  FirebaseConfig
	FCM       FCMConfig
	S3        S3Config
	Nominatim NominatimConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	AutoMigrate bool
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type FirebaseConfig struct {
	CredentialsPath string
}

type FCMConfig struct {
	CredentialsFile string
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	PublicURL string
}

type NominatimConfig struct {
	URL string
}

type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// It's ok if .env doesn't exist, we'll use environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_AUTO_MIGRATE", true)
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	viper.SetDefault("NOMINATIM_URL", "https://nominatim.openstreetmap.org")
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_DURATION", "1m")

	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		refreshExpiry = 168 * time.Hour
	}

	rateLimitDuration, err := time.ParseDuration(viper.GetString("RATE_LIMIT_DURATION"))
	if err != nil {
		rateLimitDuration = time.Minute
	}

	return &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Env:  viper.GetString("SERVER_ENV"),
		},
		Database: DatabaseConfig{
			Host:        viper.GetString("DB_HOST"),
			Port:        viper.GetString("DB_PORT"),
			User:        viper.GetString("DB_USER"),
			Password:    viper.GetString("DB_PASSWORD"),
			DBName:      viper.GetString("DB_NAME"),
			SSLMode:     viper.GetString("DB_SSL_MODE"),
			AutoMigrate: viper.GetBool("DB_AUTO_MIGRATE"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Firebase: FirebaseConfig{
			CredentialsPath: viper.GetString("FIREBASE_CREDENTIALS_PATH"),
		},
		FCM: FCMConfig{
			CredentialsFile: viper.GetString("FCM_CREDENTIALS_FILE"),
		},
		S3: S3Config{
			Endpoint:  viper.GetString("S3_ENDPOINT"),
			AccessKey: viper.GetString("S3_ACCESS_KEY_ID"),
			SecretKey: viper.GetString("S3_SECRET_ACCESS_KEY"),
			Bucket:    viper.GetString("S3_BUCKET"),
			Region:    viper.GetString("S3_REGION"),
			PublicURL: viper.GetString("S3_PUBLIC_URL"),
		},
		Nominatim: NominatimConfig{
			URL: viper.GetString("NOMINATIM_URL"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: rateLimitDuration,
		},
	}, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}
