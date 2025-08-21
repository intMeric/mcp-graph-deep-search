package scrapper

import "time"

// DefaultConfig returns a default scrapper configuration
func DefaultConfig() *Config {
	return &Config{
		UserAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout:          30 * time.Second,
		MaxContentSize:   10 * 1024 * 1024, // 10MB
		FollowRedirects:  true,
		MaxRedirects:     5,
		Parallelism:      2,
		Delay:            1 * time.Second,
		RespectRobotsTxt: true,
	}
}

// FastConfig returns a configuration optimized for speed
func FastConfig() *Config {
	return &Config{
		UserAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		Timeout:          10 * time.Second,
		MaxContentSize:   5 * 1024 * 1024, // 5MB
		FollowRedirects:  true,
		MaxRedirects:     3,
		Parallelism:      5,
		Delay:            200 * time.Millisecond,
		RespectRobotsTxt: false,
	}
}

// StealthConfig returns a configuration optimized for avoiding detection
func StealthConfig() *Config {
	return &Config{
		UserAgent:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		Timeout:          45 * time.Second,
		MaxContentSize:   15 * 1024 * 1024, // 15MB
		FollowRedirects:  true,
		MaxRedirects:     5,
		Parallelism:      1, // Very conservative
		Delay:            3 * time.Second,
		RespectRobotsTxt: true,
	}
}

// ValidateConfig validates and fixes configuration values
func ValidateConfig(config *Config) *Config {
	if config == nil {
		return DefaultConfig()
	}

	// Validate and fix values
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxContentSize <= 0 {
		config.MaxContentSize = 10 * 1024 * 1024 // 10MB
	}

	if config.MaxRedirects < 0 {
		config.MaxRedirects = 5
	}

	if config.Parallelism <= 0 {
		config.Parallelism = 2
	}

	if config.Delay < 0 {
		config.Delay = 0
	}

	// Ensure parallelism doesn't exceed reasonable limits
	if config.Parallelism > 20 {
		config.Parallelism = 20
	}

	return config
}