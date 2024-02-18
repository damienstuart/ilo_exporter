// SPDX-FileCopyrightText: (c) Damien Stuart, 2024. Licensed under [MIT](LICENSE) license.
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const version string = "1.0.3"

type WebConfig struct {
	ListenAddress string `yaml:"listen_address,omitempty"`
	MetricsPath   string `yaml:"metrics_path,omitempty"`
}

type ApiConfig struct {
	Username              string `yaml:"username,omitempty"`
	Password              string `yaml:"password,omitempty"`
	MaxConcurrentRequests uint   `yaml:"max_concurrent_requests,omitempty"`
	Debug                 bool   `yaml:"debug,omitempty"`
}

type TlsConfig struct {
	Enabled       bool   `yaml:"enabled,omitempty"`
	CertChainPath string `yaml:"cert_chain_path,omitempty"`
	KeyPath       string `yaml:"key_path,omitempty"`
}

type Config struct {
	Web WebConfig `yaml:"web,omitempty"`
	Api ApiConfig `yaml:"api,omitempty"`
	Tls TlsConfig `yaml:"tls,omitempty"`
}

var conf *Config

// DefaultConfig creates and returns the default configuration settings.
func DefaultConfig() *Config {
	defConf := Config{
		WebConfig{
			ListenAddress: ":9545",
			MetricsPath:   "/metrics",
		},
		ApiConfig{
			Username:              "",
			Password:              "",
			MaxConcurrentRequests: 4,
			Debug:                 false,
		},
		TlsConfig{
			Enabled:       false,
			CertChainPath: "",
			KeyPath:       "",
		},
	}

	return &defConf
}

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: ilo_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func printVersion() {
	fmt.Println("ilo_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Daniel Czerwonk")
	fmt.Println("Copyright: 2022, Mauve Mailorder Software GmbH & Co. KG, Licensed under MIT license")
	fmt.Println("Metric exporter for HP iLO")
}

func initConfig() {

	conf = DefaultConfig()

	showVersion := flag.Bool("version", false, "Print version information and exit.")
	configFile := flag.String("config.file", "", "The path to the configuration YAML file.")
	listenAddress := flag.String("web.listen-address", "", "Address on which to expose metrics and web interface.")
	metricsPath := flag.String("web.telemetry-path", "", "Path under which to expose metrics.")
	username := flag.String("api.username", "", "Username")
	password := flag.String("api.password", "", "Password")
	maxConcurrentRequests := flag.Uint("api.max-concurrent-requests", 0, "Maximum number of requests sent against API concurrently.")
	apiDebug := flag.Bool("api.debug", false, "Enable debugging output for API requests/responses.")
	tlsEnabled := flag.Bool("tls.enabled", false, "Enables TLS")
	tlsCertChainPath := flag.String("tls.cert-file", "", "Path to TLS cert file.")
	tlsKeyPath := flag.String("tls.key-file", "", "Path to TLS key file")

	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *configFile != "" {
		yamlFile, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to read configuration: %s\n", err)
		}
		err = yaml.Unmarshal(yamlFile, &conf)
		if err != nil {
			log.Fatalf("Error parsing YAML in config file: '%s': %s\n", *configFile, err)
		}
	}

	// Override any settings from the command-line options
	//
	if *listenAddress != "" {
		conf.Web.ListenAddress = *listenAddress
	}
	if *metricsPath != "" {
		conf.Web.MetricsPath = *metricsPath
	}
	if *username != "" {
		conf.Api.Username = *username
	}
	if *password != "" {
		conf.Api.Password = *password
	}
	if *maxConcurrentRequests != 0 {
		conf.Api.MaxConcurrentRequests = *maxConcurrentRequests
	}
	if *apiDebug {
		conf.Api.Debug = *apiDebug
	}
	if *tlsEnabled {
		conf.Tls.Enabled = *tlsEnabled
	}
	if *tlsCertChainPath != "" {
		conf.Tls.CertChainPath = *tlsCertChainPath
	}
	if *tlsKeyPath != "" {
		conf.Tls.KeyPath = *tlsKeyPath
	}
}
