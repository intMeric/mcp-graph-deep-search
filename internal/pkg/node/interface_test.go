package node

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Suite")
}

var _ = Describe("NodeConvertible Interface", func() {
	Describe("Node implementation", func() {
		var testNode *Node

		BeforeEach(func() {
			testNode = &Node{
				Type:        "TEST",
				SubType:     "example",
				DisplayName: "Test Node",
				ID:          "test-123",
			}
		})

		It("should implement NodeConvertible interface", func() {
			var convertible NodeConvertible = testNode
			Expect(convertible).NotTo(BeNil())
		})

		It("should return itself with ToNode()", func() {
			result := testNode.ToNode()
			Expect(result).To(Equal(testNode))
			Expect(result).To(BeIdenticalTo(testNode))
		})

		It("should return correct ID with GetId()", func() {
			result := testNode.GetId()
			Expect(result).To(Equal("test-123"))
		})

		It("should maintain interface contract", func() {
			var convertible NodeConvertible = testNode
			
			node := convertible.ToNode()
			id := convertible.GetId()
			
			Expect(node.ID).To(Equal(id))
			Expect(node.ID).To(Equal("test-123"))
		})
	})

	Describe("Interface behavior", func() {
		It("should work with interface polymorphism", func() {
			testNode := &Node{
				Type:        "POLY",
				DisplayName: "Polymorphic Test",
				ID:          "poly-456",
			}

			// Test using interface
			var convertible NodeConvertible = testNode
			
			node := convertible.ToNode()
			id := convertible.GetId()

			Expect(node).NotTo(BeNil())
			Expect(node.Type).To(Equal("POLY"))
			Expect(node.DisplayName).To(Equal("Polymorphic Test"))
			Expect(id).To(Equal("poly-456"))
			Expect(node.ID).To(Equal(id))
		})
	})
})