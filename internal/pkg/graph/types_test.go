package graph_test

import (
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {
	Describe("Node", func() {
		var testNode *node.Node

		BeforeEach(func() {
			testNode = &node.Node{
				Type:                "URL",
				DisplayName:         "Test URL",
				ID:                  "test-id-123",
				IsDocumentAvailable: true,
			}
		})

		Context("with valid node", func() {
			It("should validate successfully", func() {
				err := testNode.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate successfully with SubType", func() {
				testNode.SubType = "webpage"
				err := testNode.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate successfully without SubType", func() {
				testNode.SubType = ""
				err := testNode.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid node", func() {
			It("should reject empty type", func() {
				testNode.Type = ""
				err := testNode.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("type cannot be empty"))
			})

			It("should reject empty display name", func() {
				testNode.DisplayName = ""
				err := testNode.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("displayName cannot be empty"))
			})

			It("should reject whitespace-only display name", func() {
				testNode.DisplayName = "   "
				err := testNode.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("displayName cannot be empty"))
			})

			It("should reject empty ID", func() {
				testNode.ID = ""
				err := testNode.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("id cannot be empty"))
			})

			It("should reject whitespace-only ID", func() {
				testNode.ID = "   "
				err := testNode.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("id cannot be empty"))
			})
		})
	})

	Describe("Relation", func() {
		var relation *graph.Relation

		BeforeEach(func() {
			relation = &graph.Relation{
				Type:     "CONNECTED_TO",
				SourceID: "source-id-123",
				TargetID: "target-id-456",
			}
		})

		Context("with valid relation", func() {
			It("should validate successfully", func() {
				err := relation.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid relation", func() {
			It("should reject empty relation type", func() {
				relation.Type = ""
				err := relation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("relation type cannot be empty"))
			})

			It("should reject empty source ID", func() {
				relation.SourceID = ""
				err := relation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sourceId cannot be empty"))
			})

			It("should reject whitespace-only source ID", func() {
				relation.SourceID = "   "
				err := relation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sourceId cannot be empty"))
			})

			It("should reject empty target ID", func() {
				relation.TargetID = ""
				err := relation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetId cannot be empty"))
			})

			It("should reject whitespace-only target ID", func() {
				relation.TargetID = "   "
				err := relation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetId cannot be empty"))
			})

			It("should reject same source and target ID", func() {
				relation.TargetID = relation.SourceID
				err := relation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sourceId and targetId cannot be the same"))
			})
		})
	})
})
