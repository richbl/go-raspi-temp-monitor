package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Application constants
const (
	appPrefix        = "-----"
	appName          = "Go-Raspi-Temp-Monitor"
	appVersion       = "0.7.0"
	noEmailRecipient = "<none>"

	cpuTempFilePath = "/sys/class/thermal/thermal_zone0/temp"
	mailCommand     = "/usr/bin/mail"
)

// Application errors
var (
	errIsDirectory                = errors.New("'mail' command points to a directory")
	errNotExecutable              = errors.New("'mail' command is not executable")
	errTestEmailRequiresRecipient = errors.New("'-test-email' flag requires '-recipient' flag to be set")
)

// Application configuration flags
type config struct {
	EmailRecipient string
	TempThreshold  float64
	CheckInterval  time.Duration
	TestEmailFlag  bool
	Hostname       string
}

func main() {

	hello()
	cfg := parseFlags()
	cfg.Hostname = getHostname()

	if err := validateMailCommand(mailCommand); err != nil {
		log.Printf("%v", err)
		goodbye()
	}

	log.Println(appPrefix, "Configuration")
	showConfiguration(&cfg)

	// Check if -test-email flag is set
	if cfg.TestEmailFlag {
		if err := sendTestEmail(cfg); err != nil {
			log.Printf("Error sending test email: %v", err)
		}
		goodbye()
	}

	log.Println(appPrefix, "Monitoring")
	compareTemperatures(cfg) // Initial check before starting loop
	tempCheckLoop(cfg)
}

// tempCheckLoop runs the main loop to check temperature and send alerts
func tempCheckLoop(cfg config) {

	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	// Set up signal handler to monitor interrupts
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			compareTemperatures(cfg)

		case sig := <-sigChan:
			fmt.Print("\r") // Clear the ^C character from the terminal line
			log.Printf("Received signal %s: shutting down", sig)
			goodbye()
		}
	}

}

// validateMailCommand checks if the mail command is valid
func validateMailCommand(mailCommand string) error {

	info, err := os.Stat(mailCommand)
	switch {
	case errors.Is(err, os.ErrNotExist): // Check if file exists
		return fmt.Errorf("Error: '%s' %w", mailCommand, os.ErrNotExist)

	case err != nil: // Catch all other errors from os.Stat() call
		return fmt.Errorf("Error: %w", err)

	case info.IsDir(): // Check if file is a directory
		return fmt.Errorf("Error: %w '%s'", errIsDirectory, mailCommand)

	case info.Mode()&0111 == 0: // Check if file is not executable
		return fmt.Errorf("%w (%s)", errNotExecutable, mailCommand)
	}

	return nil
}

// showConfiguration displays the current configuration
func showConfiguration(cfg *config) {

	log.Printf("|\n")
	log.Printf("| Application version: %s\n", appVersion)
	log.Printf("| Temperature threshold ('-threshold'): %.2f°C\n", cfg.TempThreshold)
	log.Printf("| Check interval ('-interval'): %s\n", cfg.CheckInterval)

	if cfg.EmailRecipient == "" {
		cfg.EmailRecipient = noEmailRecipient
	}

	log.Printf("| Email recipient ('-recipient'): %s\n", cfg.EmailRecipient)
	log.Printf("| Mail command: %s\n", mailCommand)
	log.Printf("| Device hostname: %s\n", cfg.Hostname)
	log.Printf("|\n")

}

// getHostname returns the hostname of the system
func getHostname() string {

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "<unknown-host>"
	}

	return hostname
}

// parseFlags parses command line flags
func parseFlags() config {

	cfg := config{}
	flag.StringVar(&cfg.EmailRecipient, "recipient", "", "Recipient email address for alert notifications")
	flag.Float64Var(&cfg.TempThreshold, "threshold", 60.0, "CPU temperature (Celsius) threshold")
	flag.DurationVar(&cfg.CheckInterval, "interval", 5*time.Minute, "Interval for checking CPU temperature")
	flag.BoolVar(&cfg.TestEmailFlag, "test-email", false, "Send a test email and exit")
	flag.Parse()

	return cfg
}

// sendTestEmail sends a test email using the configured mail command
func sendTestEmail(cfg config) error {

	// Check if recipient is set
	if cfg.EmailRecipient == noEmailRecipient {
		return errTestEmailRequiresRecipient
	}

	// Get current CPU temperature
	currentTemp, err := getCPUTemperature()
	if err != nil {
		return err
	}

	// Create subject and body
	subject := fmt.Sprintf("%s: Test Alert (%s)", appName, cfg.Hostname)
	body := fmt.Sprintf("Warning: this is a test email\nHostname: %s\nCurrent CPU temperature: %.2f°C\nTimestamp: %s",
		cfg.Hostname, currentTemp, time.Now().Format(time.RFC1123))

	if err := sendEmail(cfg, subject, body); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// sendEmail sends an email using the configured mail command
func sendEmail(cfg config, subject, body string) error {

	if cfg.EmailRecipient == noEmailRecipient {
		log.Println("Email recipient not set: no email will be sent.")
		return nil // Not an error, just won't send
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // 15s timeout for mail command
	defer cancel()

	// Sanitize subject and recipient before passing to mail command (per GOSEC:G204)
	sanitizedSubject := strings.ReplaceAll(subject, ";", "")
	sanitizedRecipient := strings.ReplaceAll(cfg.EmailRecipient, ";", "")
	cmd := exec.CommandContext(ctx, mailCommand, "-s", sanitizedSubject, sanitizedRecipient)

	cmd.Stdin = strings.NewReader(body)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Printf("Attempting to send email to %s", cfg.EmailRecipient)

	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("failed to send email: %v", err)

		if stderr.Len() > 0 {
			errMsg += fmt.Sprintf(". Stderr: %s", stderr.String())
		}

		// Check if context deadline exceeded while sending email
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("%s (timed out)", errMsg)
			return fmt.Errorf("%s (timed out): %w", errMsg, context.DeadlineExceeded)
		}

		log.Print(errMsg)

		return fmt.Errorf("%s: %w", errMsg, err)
	}

	log.Printf("Email sent successfully to %s", cfg.EmailRecipient)

	return nil
}

// getCPUTemperature returns the current CPU temperature in Celsius
func getCPUTemperature() (float64, error) {

	data, err := os.ReadFile(cpuTempFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read temperature file %s: %w", cpuTempFilePath, err)
	}

	tempStr := strings.TrimSpace(string(data))
	tempVal, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature value '%s': %w", tempStr, err)
	}

	return tempVal / 1000.0, nil
}

// compareTemperatures checks the current temperature against the threshold and sends alerts if necessary
func compareTemperatures(cfg config) {

	currentTemp, err := getCPUTemperature()
	if err != nil {
		log.Printf("Error reading CPU temperature: %v", err)

		return
	}

	log.Printf("Current CPU temperature: %.2f°C", currentTemp)

	if currentTemp > cfg.TempThreshold {
		log.Printf("ALERT: Temperature %.2f°C exceeds threshold of %.2f°C", currentTemp, cfg.TempThreshold)

		if cfg.EmailRecipient == noEmailRecipient {
			log.Println("No recipient configured: no email notification sent")
			return
		}

		log.Println("Sending email notification")
		subject := fmt.Sprintf("%s: CPU Temp Alert (%s): %.2f°C", appName, cfg.Hostname, currentTemp)
		body := fmt.Sprintf("Warning: CPU temperature on %s has exceeded threshold\n"+
			"Threshold temp: %.2f°C\nCurrent temp: %.2f°C\nTimestamp: %s",
			cfg.Hostname, cfg.TempThreshold, currentTemp, time.Now().Format(time.RFC1123))

		if err := sendEmail(cfg, subject, body); err != nil {
			log.Printf("Error sending alert email: %v", err)
		}

	}
}

// hello outputs a welcome message
func hello() {
	log.Println(appPrefix, "Starting", appName, appVersion)
}

// goodbye outputs a goodbye message and exits the program
func goodbye() {
	log.Println(appPrefix, "Exiting", appName, appVersion)
	os.Exit(0)
}
