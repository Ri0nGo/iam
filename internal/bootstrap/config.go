package bootstrap

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Security SecurityConfig `mapstructure:"security"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	LogSQL          bool   `mapstructure:"log_sql"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Issuer        string `mapstructure:"issuer"`
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int64  `mapstructure:"expire_seconds"`
}

type SecurityConfig struct {
	PasswordCost           int `mapstructure:"password_cost"`
	LoginFailLimit         int `mapstructure:"login_fail_limit"`
	LoginFailWindowSeconds int `mapstructure:"login_fail_window_seconds"`
}

type OAuthConfig struct {
	AuthorizeCodeExpireSeconds int    `mapstructure:"authorize_code_expire_seconds"`
	LoginURL                   string `mapstructure:"login_url"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.SetEnvPrefix("IAM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if !v.IsSet("mysql.log_sql") {
		cfg.MySQL.LogSQL = true
	}
	return &cfg, nil
}
