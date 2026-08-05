package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

const defaultJWTSecret = "gophermart-default-secret"

// Config holds application settings resolved from command-line flags and
// environment variables.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
}

// Load initiate new config getting data through flags or ENV.
// Arguments must be provided or error will return.
// Flag arguments prevail on ENV arguments
func Load() (*Config, error) {
	cfg := new(Config)
	if err := parseFlag(cfg, os.Args[1:]); err != nil {
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

func parseFlag(c *Config, args []string) error {
	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&c.RunAddress, "a", "", "server address to listen on")
	fs.StringVar(&c.DatabaseURI, "d", "", "database URI")
	fs.StringVar(&c.AccrualSystemAddress, "r", "", "system address of accrual system")
	fs.StringVar(&c.JWTSecret, "j", defaultJWTSecret, "jwt secret")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("unknown flag argument provided: %v", fs.Args())
	}

	return nil
}

func parseEnv(c *Config) error {
	if c.RunAddress == "" {
		c.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if c.DatabaseURI == "" {
		c.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if c.AccrualSystemAddress == "" {
		c.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}

	if c.JWTSecret == "" {
		c.JWTSecret = os.Getenv("JWT_SECRET")
	}

	return nil
}

func validate(c *Config) error {
	if c.RunAddress == "" {
		return errors.New("no server address provided")
	}
	if c.DatabaseURI == "" {
		return errors.New("database URI is required")
	}
	if c.AccrualSystemAddress == "" {
		return errors.New("system address is required")
	}

	return nil
}
