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

func Load() *Config {
	return &Config{}
}

func (c *Config) parseFlag() error {
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

func (c *Config) parseEnv() error {
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

func (c *Config) validate() error {
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
