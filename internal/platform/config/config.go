package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	DB        DBConfig        `mapstructure:"db"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Redis     RedisConfig     `mapstructure:"redis"`
}

type ServerConfig struct {
	Port           string   `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessExpireMinute int    `mapstructure:"access_expire_minute"`
	RefreshExpireDay   int    `mapstructure:"refresh_expire_day"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type DBConfig struct {
	Driver                string `mapstructure:"driver"`
	DSN                   string `mapstructure:"dsn"`
	AutoMigrate           bool   `mapstructure:"auto_migrate"`
	MaxOpenConns          int    `mapstructure:"max_open_conns"`
	MaxIdleConns          int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinute int    `mapstructure:"conn_max_lifetime_minute"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	RPS     int  `mapstructure:"rps"`
	Burst   int  `mapstructure:"burst"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.trusted_proxies", []string{"127.0.0.1", "::1"})
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.access_expire_minute", 15)
	viper.SetDefault("jwt.refresh_expire_day", 7)
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("db.driver", "sqlite")
	viper.SetDefault("db.dsn", "app.db")
	viper.SetDefault("db.auto_migrate", true)
	viper.SetDefault("db.max_open_conns", 25)
	viper.SetDefault("db.max_idle_conns", 10)
	viper.SetDefault("db.conn_max_lifetime_minute", 15)
	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"})
	viper.SetDefault("cors.allow_headers", []string{"Content-Type", "Authorization", "X-Requested-With"})
	viper.SetDefault("cors.allow_credentials", false)
	viper.SetDefault("cors.max_age", 86400)
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.rps", 20)
	viper.SetDefault("rate_limit.burst", 40)
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "127.0.0.1")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}

	return cfg, nil
}
