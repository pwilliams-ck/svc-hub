package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "0.1.5"

type config struct {
	port   int
	env    string
	useTLS bool
	useLog bool
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cfg config

	// port defines the port number for the API server.
	// Defaults to 80 if not provided via CLI.
	flag.IntVar(&cfg.port, "port", 80, "API server port")

	// env represents the current environment the application is running in.
	// Valid values are: development, staging, production.
	// Defaults to "development" if not set via CLI.
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Boolean useTLS gives the option to enable TLS.
	// Defaults to false, use true for production.
	flag.BoolVar(&cfg.useTLS, "tls", false, "Enable TLS (true|false)")

	// Boolean useLog gives the option to enable logging to a file, as well as the usual stdout.
	// Defaults to false, use true for production.
	flag.BoolVar(&cfg.useLog, "log", false, "Enable log file (true|false)")

	// We need to parse all CLI flags in order to use them as well.
	flag.Parse()

	// Load environment variables from .envrc file.
	// err := loadEnv(".envrc")
	// if err != nil {
	// 	fmt.Println("Error loading .envrc:", err)
	// 	return
	// }

	// Logging setup
	logWriter := setupLogger(cfg.useLog)

	// Create a new logger that writes to standard output (os.Stdout).
	// Logger is configured with a text handler that formats log records as plain text.
	// nil argument specifies that no additional handler options are provided.
	logger := slog.New(slog.NewTextHandler(logWriter, nil))

	app := &application{
		config: cfg,
		logger: logger,
	}

	// TLS Config is set up for modern web , maybe remove some of these settings if needed.
	// TLS 1.3 remains unaffected by all of this, as all of its connections are considered
	// safe while writing this for Go 1.22.
	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// `srv` initializes an HTTP server with defined configuration for address, handlers,
	// TLS settings, timeouts, and error logging.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env, "tls", cfg.useTLS, "log", cfg.useLog)

	if cfg.useTLS {
		srv.TLSConfig = tlsConfig
		err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	} else {
		err := srv.ListenAndServe()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}
}

// setupLogger configures the logging output based on the useLog flag.
func setupLogger(useLog bool) io.Writer {
	var logWriter io.Writer
	if useLog {
		// Open a file for writing logs if useLog is true
		logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening log file:", err)
			os.Exit(1)
		}
		// Use a multiwriter to write logs to both standard output and the log file
		logWriter = io.MultiWriter(os.Stdout, logFile)
	} else {
		// If useLog is false, write logs only to standard output
		logWriter = os.Stdout
	}
	return logWriter
}

// Loads environment variables from a file, use typical UNIX format in .env file.
// func loadEnv(filename string) error {
// 	// Open the file
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
//
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		// Read each line
// 		line := scanner.Text()
//
// 		// Ignore comments and empty lines
// 		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
// 			continue
// 		}
//
// 		// Split the line at the first '='
// 		parts := strings.SplitN(line, "=", 2)
// 		if len(parts) != 2 {
// 			continue // skip malformed lines
// 		}
//
// 		// Set the environment variable
// 		key := parts[0]
// 		value := parts[1]
// 		os.Setenv(key, value)
// 	}
//
// 	return scanner.Err()
// }
