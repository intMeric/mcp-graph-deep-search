package scrapper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/pkg/scrapper"
)

var _ = Describe("Simple Scrapper Tests", func() {
	var (
		collyScrapper scrapper.Interface
	)

	BeforeEach(func() {
		config := scrapper.DefaultConfig()
		collyScrapper = scrapper.NewCollyScrapper(config)
	})

	AfterEach(func() {
		if collyScrapper != nil {
			collyScrapper.Close()
		}
	})

	Describe("Basic functionality", func() {
		It("should validate URLs correctly", func() {
			Expect(collyScrapper.IsValidURL("https://example.com")).To(BeTrue())
			Expect(collyScrapper.IsValidURL("http://example.com")).To(BeTrue())
			Expect(collyScrapper.IsValidURL("ftp://example.com")).To(BeFalse())
			Expect(collyScrapper.IsValidURL("")).To(BeFalse())
			Expect(collyScrapper.IsValidURL("not-a-url")).To(BeFalse())
		})

		It("should normalize URLs correctly", func() {
			normalized, err := collyScrapper.NormalizeURL("https://EXAMPLE.COM/path/")
			Expect(err).NotTo(HaveOccurred())
			Expect(normalized).To(Equal("https://example.com/path"))

			normalized, err = collyScrapper.NormalizeURL("https://example.com:443/")
			Expect(err).NotTo(HaveOccurred())
			Expect(normalized).To(Equal("https://example.com/"))
		})

		It("should handle configuration", func() {
			config := &scrapper.Config{
				UserAgent:  "TestAgent",
				Timeout:    5 * time.Second,
			}
			
			collyScrapper.SetConfig(config)
			retrievedConfig := collyScrapper.GetConfig()
			
			Expect(retrievedConfig.UserAgent).To(Equal("TestAgent"))
			Expect(retrievedConfig.Timeout).To(Equal(5 * time.Second))
		})
	})

	Describe("Configuration helpers", func() {
		It("should provide default config", func() {
			config := scrapper.DefaultConfig()
			
			Expect(config).NotTo(BeNil())
			Expect(config.UserAgent).NotTo(BeEmpty())
			Expect(config.Timeout).To(BeNumerically(">", 0))
		})

		It("should provide fast config", func() {
			config := scrapper.FastConfig()
			
			Expect(config).NotTo(BeNil())
			Expect(config.Parallelism).To(BeNumerically(">", 1))
		})

		It("should provide stealth config", func() {
			config := scrapper.StealthConfig()
			
			Expect(config).NotTo(BeNil())
			Expect(config.Parallelism).To(Equal(1))
			Expect(config.RespectRobotsTxt).To(BeTrue())
		})
	})

	Describe("Anti-detection components", func() {
		It("should create user agent rotator", func() {
			rotator := scrapper.NewUserAgentRotator()
			
			ua1 := rotator.GetRandomUserAgent()
			ua2 := rotator.GetNextUserAgent()
			
			Expect(ua1).NotTo(BeEmpty())
			Expect(ua2).NotTo(BeEmpty())
		})

		It("should create anti-detection scrapper", func() {
			config := scrapper.DefaultConfig()
			antiDetectConfig := scrapper.DefaultAntiDetectionConfig()
			
			ads := scrapper.NewAntiDetectionScrapper(config, antiDetectConfig)
			
			Expect(ads).NotTo(BeNil())
			Expect(ads.GetConfig()).NotTo(BeNil())
			Expect(ads.GetAntiDetectionConfig()).NotTo(BeNil())
		})
	})
})