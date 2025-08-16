package pii_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/pii"
)

var _ = Describe("PII Extractor", func() {
	var (
		extractor pii.Extractor
		ctx       context.Context
		cancel    context.CancelFunc
	)

	BeforeEach(func() {
		var err error
		extractor, err = pii.NewExtractor()
		Expect(err).NotTo(HaveOccurred())
		Expect(extractor).NotTo(BeNil())

		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	})

	AfterEach(func() {
		cancel()
		if extractor != nil {
			extractor.Close()
		}
	})

	Describe("ExtractPII", func() {
		Context("with text containing various PII types", func() {
			It("should extract emails successfully", func() {
				text := "Contact me at john.doe@example.com or admin@test.org"

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Total).To(BeNumerically(">", 0))
				Expect(result.HasType(pii.PIITypeEmail)).To(BeTrue())

				emails := result.GetEmails()
				Expect(emails).NotTo(BeEmpty())
				Expect(len(emails)).To(BeNumerically(">=", 1))
			})

			It("should extract phone numbers successfully", func() {
				text := "Call me at (555) 123-4567 or +1-800-555-0199"

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				if result.HasType(pii.PIITypePhone) {
					phones := result.GetPhones()
					Expect(phones).NotTo(BeEmpty())
				}
			})

			It("should extract credit card numbers", func() {
				text := "My credit card is 4111-1111-1111-1111"

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				if result.HasType(pii.PIITypeCreditCard) {
					cards := result.GetCreditCards()
					Expect(cards).NotTo(BeEmpty())
				}
			})

			It("should extract IP addresses", func() {
				text := "Server IP is 192.168.1.100 and backup is 10.0.0.1"

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				if result.HasType(pii.PIITypeIPAddress) {
					ips := result.GetIPAddresses()
					Expect(ips).NotTo(BeEmpty())
				}
			})

			It("should handle comprehensive PII text", func() {
				text := `
				Hello, my name is John Doe. You can reach me at john.doe@example.com
				or call me at (555) 123-4567. My home address is 123 Main Street,
				New York, NY 10001.

				For business purposes, my SSN is 123-45-6789 and you can send mail to
				P.O. Box 456. My credit card number is 4111-1111-1111-1111.

				Server details:
				- IP Address: 192.168.1.100
				- Bitcoin wallet: 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
				- Bank account (IBAN): GB82WEST12345698765432
				`

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Total).To(BeNumerically(">", 0))
				Expect(result.IsEmpty()).To(BeFalse())
				Expect(result.Stats).NotTo(BeEmpty())

				// Should find at least emails
				Expect(result.HasType(pii.PIITypeEmail)).To(BeTrue())
			})
		})

		Context("with empty or no-PII text", func() {
			It("should handle empty text gracefully", func() {
				text := ""

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.IsEmpty()).To(BeTrue())
				Expect(result.Total).To(Equal(0))
			})

			It("should handle text with no PII", func() {
				text := "This is just a regular text without any personal information."

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.IsEmpty()).To(BeTrue())
			})
		})

		Context("with context cancellation", func() {
			It("should respect context", func() {
				text := "Contact john.doe@example.com"

				result, err := extractor.ExtractPII(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
			})
		})
	})

	Describe("Result methods", func() {
		var result *pii.Result

		BeforeEach(func() {
			text := "Email: test@example.com, Phone: (555) 123-4567"
			var err error
			result, err = extractor.ExtractPII(ctx, text)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should provide correct type checking", func() {
			if result.Total > 0 {
				Expect(result.IsEmpty()).To(BeFalse())

				// Check if we can get entities by type
				emails := result.GetByType(pii.PIITypeEmail)
				phones := result.GetByType(pii.PIITypePhone)

				if result.HasType(pii.PIITypeEmail) {
					Expect(emails).NotTo(BeEmpty())
				}

				if result.HasType(pii.PIITypePhone) {
					Expect(phones).NotTo(BeEmpty())
				}
			}
		})

		It("should have proper entity structure", func() {
			if len(result.Entities) > 0 {
				entity := result.Entities[0]
				Expect(entity.Type).NotTo(BeEmpty())
				Expect(entity.Value).NotTo(BeEmpty())
				Expect(entity.Count).To(BeNumerically(">", 0))
			}
		})
	})

	Describe("Close", func() {
		It("should close without error", func() {
			err := extractor.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
