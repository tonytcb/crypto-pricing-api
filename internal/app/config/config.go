package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

const (
	mainEnvFile = ".env"
	hide        = "hide"
)

type Config struct {
	Environment string `mapstructure:"ENV"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	RestAPIPort string `mapstructure:"REST_API_PORT"`

	PairPriceToMonitor        string        `mapstructure:"PAIR_PRICE_TO_MONITOR"`
	StoreMaxItems             int           `mapstructure:"STORE_MAX_ITEMS"`
	SseClientsBufferSize      int           `mapstructure:"SSE_CLIENTS_BUFFER_SIZE"`
	SSEClientsCleanUpInterval time.Duration `mapstructure:"SSE_CLIENTS_CLEAN_UP_INTERVAL"`

	// CoinDesk HTTP Client configurations
	CoinDeskAPIURL           string        `mapstructure:"COIN_DESK_API_URL"`
	CoinDeskRetryMaxAttempts int           `mapstructure:"COIN_DESK_RETRY_MAX_ATTEMPTS"`
	CoinDeskClientTimeout    time.Duration `mapstructure:"COIN_DESK_CLIENT_TIMEOUT"`
	CoinDeskRetryTimeout     time.Duration `mapstructure:"COIN_DESK_RETRY_TIMEOUT"`
	CoinDeskRetryInitialWait time.Duration `mapstructure:"COIN_DESK_RETRY_INITIAL_WAIT"`
	CoinDeskRetryMaxWait     time.Duration `mapstructure:"COIN_DESK_RETRY_MAX_WAIT"`

	// Pulling configurations
	PricesPullingInterval   time.Duration `mapstructure:"PRICES_PULLING_INTERVAL"`
	PricesChannelBufferSize int           `mapstructure:"PRICES_CHANNEL_BUFFER_SIZE"`
}

func (c Config) PairToMonitor() (domain.Pair, error) {
	return domain.NewPairFromString(c.PairPriceToMonitor)
}

func Load(filenames ...string) (*Config, error) {
	var cfg = &Config{}

	filenames = append(filenames, mainEnvFile)

	viper.SetConfigType("env")
	viper.AutomaticEnv()

	for _, filename := range filenames {
		if _, err := os.Stat(filename); err != nil {
			continue
		}

		viper.SetConfigFile(filename)

		if err := viper.ReadInConfig(); err != nil {
			return nil, errors.Wrapf(err, "error to read config, path: %s", mainEnvFile)
		}

		if err := viper.MergeInConfig(); err != nil {
			return nil, errors.Wrapf(err, "error to merge config, filename: %s", filename)
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			return nil, errors.Wrapf(err, "error to unmarshal config, filename: %s", filename)
		}
	}

	return cfg, nil
}
