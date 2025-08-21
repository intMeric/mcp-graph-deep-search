package scrapper

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// UserAgentRotator manages user agent rotation
type UserAgentRotator struct {
	userAgents []string
	current    int
	mutex      sync.RWMutex
}

// NewUserAgentRotator creates a new user agent rotator with common browser user agents
func NewUserAgentRotator() *UserAgentRotator {
	return &UserAgentRotator{
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		},
	}
}

// GetRandomUserAgent returns a random user agent
func (uar *UserAgentRotator) GetRandomUserAgent() string {
	uar.mutex.RLock()
	defer uar.mutex.RUnlock()
	
	if len(uar.userAgents) == 0 {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	
	return uar.userAgents[rand.Intn(len(uar.userAgents))]
}

// GetNextUserAgent returns the next user agent in rotation
func (uar *UserAgentRotator) GetNextUserAgent() string {
	uar.mutex.Lock()
	defer uar.mutex.Unlock()
	
	if len(uar.userAgents) == 0 {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	
	userAgent := uar.userAgents[uar.current]
	uar.current = (uar.current + 1) % len(uar.userAgents)
	return userAgent
}

// AddUserAgent adds a new user agent to the rotation
func (uar *UserAgentRotator) AddUserAgent(userAgent string) {
	uar.mutex.Lock()
	defer uar.mutex.Unlock()
	
	uar.userAgents = append(uar.userAgents, userAgent)
}


// AntiDetectionConfig holds anti-detection settings
type AntiDetectionConfig struct {
	EnableUserAgentRotation   bool          `json:"enableUserAgentRotation"`
	EnableDelayRandomization  bool          `json:"enableDelayRandomization"`
	MinDelay                  time.Duration `json:"minDelay"`
	MaxDelay                  time.Duration `json:"maxDelay"`
	EnableHeaderRandomization bool          `json:"enableHeaderRandomization"`
}

// DefaultAntiDetectionConfig returns a default anti-detection configuration
func DefaultAntiDetectionConfig() *AntiDetectionConfig {
	return &AntiDetectionConfig{
		EnableUserAgentRotation:   true,
		EnableDelayRandomization:  true,
		MinDelay:                  500 * time.Millisecond,
		MaxDelay:                  2 * time.Second,
		EnableHeaderRandomization: true,
	}
}

// AntiDetectionScrapper extends CollyScrapper with anti-detection capabilities
type AntiDetectionScrapper struct {
	*CollyScrapper
	userAgentRotator *UserAgentRotator
	antiDetectConfig *AntiDetectionConfig
}

// NewAntiDetectionScrapper creates a scrapper with anti-detection features
func NewAntiDetectionScrapper(config *Config, antiDetectConfig *AntiDetectionConfig) *AntiDetectionScrapper {
	if antiDetectConfig == nil {
		antiDetectConfig = DefaultAntiDetectionConfig()
	}

	baseScrapper := NewCollyScrapper(config).(*CollyScrapper)
	
	ads := &AntiDetectionScrapper{
		CollyScrapper:    baseScrapper,
		userAgentRotator: NewUserAgentRotator(),
		antiDetectConfig: antiDetectConfig,
	}

	ads.setupAntiDetection()
	return ads
}


// setupAntiDetection configures the collector with anti-detection features
func (ads *AntiDetectionScrapper) setupAntiDetection() {
	c := ads.collector

	// Setup user agent and header rotation
	c.OnRequest(func(r *colly.Request) {
		// Random user agent
		if ads.antiDetectConfig.EnableUserAgentRotation {
			userAgent := ads.userAgentRotator.GetRandomUserAgent()
			r.Headers.Set("User-Agent", userAgent)
		}

		// Setup realistic headers
		if ads.antiDetectConfig.EnableHeaderRandomization {
			ads.setRealisticHeaders(r)
		}

		// Random delay
		if ads.antiDetectConfig.EnableDelayRandomization {
			delay := ads.getRandomDelay()
			time.Sleep(delay)
		}
	})


}

// setRealisticHeaders sets realistic browser headers
func (ads *AntiDetectionScrapper) setRealisticHeaders(r *colly.Request) {
	headers := [][]string{
		{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		{"Accept-Language", "en-US,en;q=0.9"},
		{"Accept-Encoding", "gzip, deflate, br"},
		{"Cache-Control", "max-age=0"},
		{"Connection", "keep-alive"},
		{"Upgrade-Insecure-Requests", "1"},
		{"Sec-Fetch-Dest", "document"},
		{"Sec-Fetch-Mode", "navigate"},
		{"Sec-Fetch-Site", "none"},
		{"Sec-Fetch-User", "?1"},
	}

	// Add some headers randomly to make requests more varied
	for _, header := range headers {
		if rand.Float32() < 0.8 { // 80% chance to include each header
			r.Headers.Set(header[0], header[1])
		}
	}

	// Random DNT header
	if rand.Float32() < 0.3 { // 30% chance
		r.Headers.Set("DNT", "1")
	}
}

// getRandomDelay returns a random delay between min and max
func (ads *AntiDetectionScrapper) getRandomDelay() time.Duration {
	if ads.antiDetectConfig.MinDelay >= ads.antiDetectConfig.MaxDelay {
		return ads.antiDetectConfig.MinDelay
	}

	diff := ads.antiDetectConfig.MaxDelay - ads.antiDetectConfig.MinDelay
	randomDiff := time.Duration(rand.Int63n(int64(diff)))
	return ads.antiDetectConfig.MinDelay + randomDiff
}


// UpdateAntiDetectionConfig updates the anti-detection configuration
func (ads *AntiDetectionScrapper) UpdateAntiDetectionConfig(config *AntiDetectionConfig) {
	ads.antiDetectConfig = config
	ads.setupAntiDetection()
}

// GetAntiDetectionConfig returns the current anti-detection configuration
func (ads *AntiDetectionScrapper) GetAntiDetectionConfig() *AntiDetectionConfig {
	return ads.antiDetectConfig
}