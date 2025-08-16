package pii_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/pii"
)

var _ = Describe("Interface", func() {
	Describe("Result", func() {
		var result *pii.Result

		BeforeEach(func() {
			result = &pii.Result{
				Total: 3,
				Entities: []pii.Entity{
					{
						Type:     pii.PIITypeEmail,
						Value:    "test@example.com",
						Count:    1,
						Contexts: []string{"Contact test@example.com"},
					},
					{
						Type:     pii.PIITypePhone,
						Value:    "(555) 123-4567",
						Count:    1,
						Contexts: []string{"Call (555) 123-4567"},
					},
					{
						Type:     pii.PIITypeEmail,
						Value:    "admin@test.org",
						Count:    1,
						Contexts: []string{"Email admin@test.org"},
					},
				},
				Stats: map[string]int{
					"email": 2,
					"phone": 1,
				},
			}
		})

		It("should report correct emptiness", func() {
			Expect(result.IsEmpty()).To(BeFalse())

			emptyResult := &pii.Result{Total: 0}
			Expect(emptyResult.IsEmpty()).To(BeTrue())
		})

		It("should check type existence correctly", func() {
			Expect(result.HasType(pii.PIITypeEmail)).To(BeTrue())
			Expect(result.HasType(pii.PIITypePhone)).To(BeTrue())
			Expect(result.HasType(pii.PIITypeCreditCard)).To(BeFalse())
		})

		It("should filter entities by type", func() {
			emails := result.GetByType(pii.PIITypeEmail)
			Expect(emails).To(HaveLen(2))

			phones := result.GetByType(pii.PIITypePhone)
			Expect(phones).To(HaveLen(1))

			creditCards := result.GetByType(pii.PIITypeCreditCard)
			Expect(creditCards).To(BeEmpty())
		})

		It("should provide convenience methods", func() {
			emails := result.GetEmails()
			Expect(emails).To(HaveLen(2))

			phones := result.GetPhones()
			Expect(phones).To(HaveLen(1))

			creditCards := result.GetCreditCards()
			Expect(creditCards).To(BeEmpty())

			ips := result.GetIPAddresses()
			Expect(ips).To(BeEmpty())
		})
	})

	Describe("Entity", func() {
		It("should have proper structure", func() {
			entity := pii.Entity{
				Type:     pii.PIITypeEmail,
				Value:    "test@example.com",
				Count:    2,
				Contexts: []string{"Email: test@example.com", "Contact test@example.com"},
			}

			Expect(entity.Type).To(Equal(pii.PIITypeEmail))
			Expect(entity.Value).To(Equal("test@example.com"))
			Expect(entity.Count).To(Equal(2))
			Expect(entity.Contexts).To(HaveLen(2))
		})

		Describe("ToNode", func() {
			It("should convert entity to node correctly", func() {
				entity := pii.Entity{
					Type:     pii.PIITypeEmail,
					Value:    "test@example.com",
					Count:    1,
					Contexts: []string{"Contact test@example.com"},
				}

				node := entity.ToNode()

				Expect(node).NotTo(BeNil())
				Expect(node.Type).To(Equal("PII"))
				Expect(node.SubType).To(Equal("email"))
				Expect(node.DisplayName).To(Equal("email: test@example.com"))
				Expect(node.ID).To(HavePrefix("pii-email-"))
				Expect(len(node.ID)).To(BeNumerically(">", 10))
				Expect(node.Location).To(BeEmpty())
			})

			It("should implement NodeConvertible interface", func() {
				entity := pii.Entity{
					Type:  pii.PIITypePhone,
					Value: "(555) 123-4567",
				}

				id := entity.GetId()
				node := entity.ToNode()

				Expect(id).To(Equal(node.ID))
				Expect(id).To(HavePrefix("pii-phone-"))
			})

			It("should generate consistent IDs for same values", func() {
				entity1 := pii.Entity{
					Type:  pii.PIITypePhone,
					Value: "(555) 123-4567",
				}
				entity2 := pii.Entity{
					Type:  pii.PIITypePhone,
					Value: "(555) 123-4567",
				}

				node1 := entity1.ToNode()
				node2 := entity2.ToNode()

				Expect(node1.ID).To(Equal(node2.ID))
				Expect(node1.DisplayName).To(Equal(node2.DisplayName))
				Expect(node1.SubType).To(Equal("phone"))
				Expect(node2.SubType).To(Equal("phone"))
			})

			It("should generate different IDs for different values", func() {
				entity1 := pii.Entity{
					Type:  pii.PIITypeEmail,
					Value: "user1@example.com",
				}
				entity2 := pii.Entity{
					Type:  pii.PIITypeEmail,
					Value: "user2@example.com",
				}

				node1 := entity1.ToNode()
				node2 := entity2.ToNode()

				Expect(node1.ID).NotTo(Equal(node2.ID))
				Expect(node1.DisplayName).NotTo(Equal(node2.DisplayName))
			})

			It("should create valid nodes that pass validation", func() {
				entity := pii.Entity{
					Type:  pii.PIITypeCreditCard,
					Value: "1234-5678-9012-3456",
				}

				node := entity.ToNode()

				err := node.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("PIIType constants", func() {
		It("should have all expected types", func() {
			Expect(pii.PIITypeEmail).To(Equal(pii.PIIType("email")))
			Expect(pii.PIITypePhone).To(Equal(pii.PIIType("phone")))
			Expect(pii.PIITypeCreditCard).To(Equal(pii.PIIType("credit_card")))
			Expect(pii.PIITypeSSN).To(Equal(pii.PIIType("ssn")))
			Expect(pii.PIITypeIPAddress).To(Equal(pii.PIIType("ip_address")))
			Expect(pii.PIITypeAddress).To(Equal(pii.PIIType("address")))
			Expect(pii.PIITypeName).To(Equal(pii.PIIType("name")))
			Expect(pii.PIITypeBitcoin).To(Equal(pii.PIIType("bitcoin")))
			Expect(pii.PIITypeIBAN).To(Equal(pii.PIIType("iban")))
			Expect(pii.PIITypeOther).To(Equal(pii.PIIType("other")))
		})
	})
})
