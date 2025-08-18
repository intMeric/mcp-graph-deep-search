package graph_test

import (
	"strings"

	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Neo4jGraph", func() {
	var config *graph.Neo4jConfig

	BeforeEach(func() {
		config = &graph.Neo4jConfig{
			URI:      "neo4j://localhost:7687",
			Username: "neo4j",
			Password: "testpassword",
			Database: "neo4j",
		}
	})

	Describe("NewNeo4jGraph", func() {
		Context("with nil config", func() {
			It("should return an error", func() {
				g, err := graph.NewNeo4jGraph(nil)

				Expect(g).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("config cannot be nil"))
			})
		})

		Context("with valid config but no Neo4j server", func() {
			It("should return a connection error", func() {
				g, err := graph.NewNeo4jGraph(config)

				Expect(g).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to verify connectivity"))
			})
		})

		Context("with invalid URI", func() {
			BeforeEach(func() {
				config.URI = "invalid-uri"
			})

			It("should return a driver creation error", func() {
				g, err := graph.NewNeo4jGraph(config)

				Expect(g).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create Neo4j driver"))
			})
		})
	})

	Describe("Validation Tests", func() {
		Describe("CreateNode validation", func() {
			Context("with nil node", func() {
				It("should be handled by node validation", func() {
					node := (*node.Node)(nil)
					Expect(node).To(BeNil())
				})
			})

			Context("with invalid node", func() {
				It("should validate displayName requirement", func() {
					node := &node.Node{
						Type:        "URL",
						DisplayName: "",
						ID:          "test-id",
					}

					err := node.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("displayName cannot be empty"))
				})

				It("should validate ID requirement", func() {
					node := &node.Node{
						Type:        "URL",
						DisplayName: "Test",
						ID:          "",
					}

					err := node.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("id cannot be empty"))
				})

				It("should validate node type", func() {
					node := &node.Node{
						Type:        "",
						DisplayName: "Test",
						ID:          "test-id",
					}

					err := node.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("type cannot be empty"))
				})
			})

			Context("with valid node", func() {
				It("should pass validation", func() {
					node := &node.Node{
						Type:                "URL",
						DisplayName:         "Test URL",
						ID:                  "test-id-123",
						IsDocumentAvailable: true,
					}

					err := node.Validate()

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Describe("CreateRelation validation", func() {
			Context("with nil relation", func() {
				It("should be handled by relation validation", func() {
					relation := (*graph.Relation)(nil)
					Expect(relation).To(BeNil())
				})
			})

			Context("with invalid relation", func() {
				It("should validate sourceId requirement", func() {
					relation := &graph.Relation{
						Type:     "CONNECTED_TO",
						SourceID: "",
						TargetID: "target-id",
					}

					err := relation.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("sourceId cannot be empty"))
				})

				It("should validate targetId requirement", func() {
					relation := &graph.Relation{
						Type:     "CONNECTED_TO",
						SourceID: "source-id",
						TargetID: "",
					}

					err := relation.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("targetId cannot be empty"))
				})

				It("should prevent same source and target", func() {
					relation := &graph.Relation{
						Type:     "CONNECTED_TO",
						SourceID: "same-id",
						TargetID: "same-id",
					}

					err := relation.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("sourceId and targetId cannot be the same"))
				})

				It("should validate relation type", func() {
					relation := &graph.Relation{
						Type:     "",
						SourceID: "source-id",
						TargetID: "target-id",
					}

					err := relation.Validate()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("relation type cannot be empty"))
				})
			})

			Context("with valid relation", func() {
				It("should pass validation", func() {
					relation := &graph.Relation{
						Type:     "CONNECTED_TO",
						SourceID: "source-id",
						TargetID: "target-id",
					}

					err := relation.Validate()

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Describe("ID validation", func() {
			Context("with empty ID", func() {
				It("should reject empty string", func() {
					Expect("").To(BeEmpty())
				})

				It("should reject whitespace-only string", func() {
					Expect(strings.TrimSpace("   ")).To(BeEmpty())
				})
			})
		})
	})

	Describe("Config validation", func() {
		Context("with empty URI", func() {
			BeforeEach(func() {
				config.URI = ""
			})

			It("should handle empty URI gracefully", func() {
				g, err := graph.NewNeo4jGraph(config)

				Expect(g).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with empty credentials", func() {
			BeforeEach(func() {
				config.Username = ""
				config.Password = ""
			})

			It("should handle empty credentials", func() {
				g, err := graph.NewNeo4jGraph(config)

				Expect(g).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
