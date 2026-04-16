package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	S3       S3Config
	SMTP     SMTPConfig
	SMS      SMSConfig
	Payment  PaymentConfig
	CORS     CORSConfig
	OTel     OTelConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Env  string `mapstructure:"APP_ENV"`
	Port string `mapstructure:"APP_PORT"`
}

// DBConfig holds database connection settings.
type DBConfig struct {
	DSN      string `mapstructure:"DB_DSN"`
	MaxConns int32  `mapstructure:"DB_MAX_CONNS"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	URL string `mapstructure:"REDIS_URL"`
}

// JWTConfig holds JWT token settings.
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"JWT_ACCESS_SECRET"`
	RefreshSecret string        `mapstructure:"JWT_REFRESH_SECRET"`
	AccessTTL     time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	RefreshTTL    time.Duration `mapstructure:"JWT_REFRESH_TTL"`
}

// S3Config holds S3/R2 storage settings.
type S3Config struct {
	Bucket          string `mapstructure:"S3_BUCKET"`
	Region          string `mapstructure:"S3_REGION"`
	Endpoint        string `mapstructure:"S3_ENDPOINT"`
	AccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
}

// SMTPConfig holds email sending settings.
type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	From     string `mapstructure:"SMTP_FROM"`
}

// SMSConfig holds SMS provider settings.
type SMSConfig struct {
	APIKey   string `mapstructure:"AT_API_KEY"`
	Username string `mapstructure:"AT_USERNAME"`
	SenderID string `mapstructure:"AT_SENDER_ID"`
}

// PaymentConfig holds payment provider settings.
type PaymentConfig struct {
	MTN    MTNConfig
	Orange OrangeConfig
}

// MTNConfig holds MTN MoMo settings.
type MTNConfig struct {
	APIKey      string `mapstructure:"MTN_MOMO_API_KEY"`
	APISecret   string `mapstructure:"MTN_MOMO_API_SECRET"`
	BaseURL     string `mapstructure:"MTN_MOMO_BASE_URL"`
	CallbackURL string `mapstructure:"MTN_MOMO_CALLBACK_URL"`
}

// OrangeConfig holds Orange Money settings.
type OrangeConfig struct {
	APIKey      string `mapstructure:"ORANGE_MONEY_API_KEY"`
	APISecret   string `mapstructure:"ORANGE_MONEY_API_SECRET"`
	BaseURL     string `mapstructure:"ORANGE_MONEY_BASE_URL"`
	CallbackURL string `mapstructure:"ORANGE_MONEY_CALLBACK_URL"`
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	Origins []string
}

// OTelConfig holds OpenTelemetry settings.
type OTelConfig struct {
	ExporterEndpoint string `mapstructure:"OTEL_EXPORTER_ENDPOINT"`
}

// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8000")
	viper.SetDefault("DB_MAX_CONNS", 25)
	viper.SetDefault("JWT_ACCESS_TTL", "15m")
	viper.SetDefault("JWT_REFRESH_TTL", "168h")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("AT_USERNAME", "sandbox")
	viper.SetDefault("AT_SENDER_ID", "UrbanSanc")
	viper.SetDefault("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173")
	viper.SetDefault("OTEL_EXPORTER_ENDPOINT", "localhost:4317")

	// Try reading .env file, ignore error if not found
	_ = viper.ReadInConfig()

	accessTTL, err := time.ParseDuration(viper.GetString("JWT_ACCESS_TTL"))
	if err != nil {
		accessTTL = 15 * time.Minute
	}

	refreshTTL, err := time.ParseDuration(viper.GetString("JWT_REFRESH_TTL"))
	if err != nil {
		refreshTTL = 7 * 24 * time.Hour
	}

	corsOrigins := strings.Split(viper.GetString("CORS_ORIGINS"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	cfg := &Config{
		App: AppConfig{
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetString("APP_PORT"),
		},
		DB: DBConfig{
			DSN:      viper.GetString("DB_DSN"),
			MaxConns: viper.GetInt32("DB_MAX_CONNS"),
		},
		Redis: RedisConfig{
			URL: viper.GetString("REDIS_URL"),
		},
		JWT: JWTConfig{
			AccessSecret:  viper.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret: viper.GetString("JWT_REFRESH_SECRET"),
			AccessTTL:     accessTTL,
			RefreshTTL:    refreshTTL,
		},
		S3: S3Config{
			Bucket:          viper.GetString("S3_BUCKET"),
			Region:          viper.GetString("S3_REGION"),
			Endpoint:        viper.GetString("S3_ENDPOINT"),
			AccessKeyID:     viper.GetString("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: viper.GetString("AWS_SECRET_ACCESS_KEY"),
		},
		SMTP: SMTPConfig{
			Host:     viper.GetString("SMTP_HOST"),
			Port:     viper.GetInt("SMTP_PORT"),
			Username: viper.GetString("SMTP_USERNAME"),
			Password: viper.GetString("SMTP_PASSWORD"),
			From:     viper.GetString("SMTP_FROM"),
		},
		SMS: SMSConfig{
			APIKey:   viper.GetString("AT_API_KEY"),
			Username: viper.GetString("AT_USERNAME"),
			SenderID: viper.GetString("AT_SENDER_ID"),
		},
		Payment: PaymentConfig{
			MTN: MTNConfig{
				APIKey:      viper.GetString("MTN_MOMO_API_KEY"),
				APISecret:   viper.GetString("MTN_MOMO_API_SECRET"),
				BaseURL:     viper.GetString("MTN_MOMO_BASE_URL"),
				CallbackURL: viper.GetString("MTN_MOMO_CALLBACK_URL"),
			},
			Orange: OrangeConfig{
				APIKey:      viper.GetString("ORANGE_MONEY_API_KEY"),
				APISecret:   viper.GetString("ORANGE_MONEY_API_SECRET"),
				BaseURL:     viper.GetString("ORANGE_MONEY_BASE_URL"),
				CallbackURL: viper.GetString("ORANGE_MONEY_CALLBACK_URL"),
			},
		},
		CORS: CORSConfig{
			Origins: corsOrigins,
		},
		OTel: OTelConfig{
			ExporterEndpoint: viper.GetString("OTEL_EXPORTER_ENDPOINT"),
		},
	}

	return cfg, nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
