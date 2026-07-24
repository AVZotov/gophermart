package internal

import (
	"errors"
	"flag"
	"log"
	"os"
)

type Config struct {
	RunAddress           string
	DataBaseURI          string
	AccrualSystemAddress string
}

// Load initiate new config getting data through flags or ENV.
// Arguments must be provided or error will return.
// Flag arguments prevail on ENV arguments
func Load() (*Config, error) {
	cfg := new(Config)
	if err := parseFlag(cfg); err != nil {
		return nil, err
	}
	if err := parseEnv(cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseFlag(c *Config) error {
	flag.StringVar(&c.RunAddress, "a", "", "server address to listen on")
	flag.StringVar(&c.DataBaseURI, "d", "", "database URI")
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "system address of accrual system")

	flag.Parse()

	if flag.NArg() > 0 {
		for _, arg := range flag.Args() {
			log.Printf("unknown argument: %s\n", arg)
		}
		flag.Usage()
		return errors.New("unknown flag argument provided")
	}

	return nil
}

func parseEnv(c *Config) error {
	if c.RunAddress == "" {
		c.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if c.DataBaseURI == "" {
		c.DataBaseURI = os.Getenv("DATABASE_URI")
	}
	if c.AccrualSystemAddress == "" {
		c.AccrualSystemAddress = os.Getenv("ACCURAL_SYSTEM_ADDRESS")
	}
	return nil
}

func validate(c *Config) error {
	if c.RunAddress == "" {
		return errors.New("no server address provided")
	}
	if c.DataBaseURI == "" {
		return errors.New("database URI is required")
	}
	if c.AccrualSystemAddress == "" {
		return errors.New("system address is required")
	}

	return nil
}
