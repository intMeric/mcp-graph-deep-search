package object_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/pkg/node"
	"mgds/pkg/object"
)

var _ = Describe("URLNode", func() {
	var (
		urlNode node.Interface
		mgdsId  string
		name    string
		desc    string
		url     string
	)

	BeforeEach(func() {
		mgdsId = "test-id-123"
		name = "Test Page"
		desc = "A test web page"
		url = "https://example.com"
	})

	Describe("NewURLNode", func() {
		Context("with valid parameters", func() {
			It("should create a valid URL node", func() {
				urlNode = object.NewURLNode(mgdsId, name, desc, url)

				Expect(urlNode).NotTo(BeNil())
				Expect(urlNode.GetMgdsId()).To(Equal(mgdsId))
				Expect(urlNode.GetDisplayName()).To(Equal(name))
				Expect(urlNode.GetDescription()).To(Equal(desc))
				Expect(urlNode.GetType()).To(Equal(object.URLNodeType))
				Expect(urlNode.IsValid()).To(BeTrue())
			})
		})

		Context("with empty parameters", func() {
			It("should create an invalid node with empty mgdsId", func() {
				urlNode = object.NewURLNode("", name, desc, url)
				Expect(urlNode.IsValid()).To(BeFalse())
			})

			It("should create an invalid node with empty displayName", func() {
				urlNode = object.NewURLNode(mgdsId, "", desc, url)
				Expect(urlNode.IsValid()).To(BeFalse())
			})

			It("should create an invalid node with empty description", func() {
				urlNode = object.NewURLNode(mgdsId, name, "", url)
				Expect(urlNode.IsValid()).To(BeFalse())
			})

			It("should create an invalid node with empty URL", func() {
				urlNode = object.NewURLNode(mgdsId, name, desc, "")
				Expect(urlNode.IsValid()).To(BeFalse())
			})
		})
	})

	Describe("NewURLNodeWithDetails", func() {
		Context("with valid parameters including title and content", func() {
			It("should create a URL node with additional details", func() {
				title := "Example Title"
				content := "Example content of the page"
				
				urlNode = object.NewURLNodeWithDetails(mgdsId, name, desc, url, title, content)
				urlNodeConcrete := urlNode.(*object.URLNode)

				Expect(urlNode).NotTo(BeNil())
				Expect(urlNode.IsValid()).To(BeTrue())
				Expect(urlNodeConcrete.GetTitle()).To(Equal(title))
				Expect(urlNodeConcrete.GetContent()).To(Equal(content))
			})
		})
	})

	Describe("Properties Management", func() {
		BeforeEach(func() {
			urlNode = object.NewURLNode(mgdsId, name, desc, url)
		})

		Context("setting and getting properties", func() {
			It("should handle string properties", func() {
				key := "author"
				value := "John Doe"
				
				urlNode.SetProperty(key, value)
				retrievedValue, exists := urlNode.GetProperty(key)
				
				Expect(exists).To(BeTrue())
				Expect(retrievedValue).To(Equal(value))
			})

			It("should handle complex properties", func() {
				key := "metadata"
				value := map[string]string{"lang": "en", "charset": "utf-8"}
				
				urlNode.SetProperty(key, value)
				retrievedValue, exists := urlNode.GetProperty(key)
				
				Expect(exists).To(BeTrue())
				Expect(retrievedValue).To(Equal(value))
			})

			It("should return false for non-existent properties", func() {
				value, exists := urlNode.GetProperty("non-existent")
				
				Expect(exists).To(BeFalse())
				Expect(value).To(BeNil())
			})
		})
	})

	Describe("URL-specific methods", func() {
		var urlNodeConcrete *object.URLNode

		BeforeEach(func() {
			urlNode = object.NewURLNode(mgdsId, name, desc, url)
			urlNodeConcrete = urlNode.(*object.URLNode)
		})

		Context("URL management", func() {
			It("should get and set URL", func() {
				newURL := "https://newexample.com"
				originalTime := urlNodeConcrete.GetUpdatedAt()
				
				time.Sleep(time.Millisecond) // Ensure time difference
				urlNodeConcrete.SetURL(newURL)
				
				Expect(urlNodeConcrete.GetURL()).To(Equal(newURL))
				Expect(urlNodeConcrete.GetUpdatedAt()).To(BeTemporally(">", originalTime))
			})
		})

		Context("Title management", func() {
			It("should get and set title", func() {
				newTitle := "New Page Title"
				originalTime := urlNodeConcrete.GetUpdatedAt()
				
				time.Sleep(time.Millisecond)
				urlNodeConcrete.SetTitle(newTitle)
				
				Expect(urlNodeConcrete.GetTitle()).To(Equal(newTitle))
				Expect(urlNodeConcrete.GetUpdatedAt()).To(BeTemporally(">", originalTime))
			})
		})

		Context("Content management", func() {
			It("should get and set content", func() {
				newContent := "This is the new page content"
				originalTime := urlNodeConcrete.GetUpdatedAt()
				
				time.Sleep(time.Millisecond)
				urlNodeConcrete.SetContent(newContent)
				
				Expect(urlNodeConcrete.GetContent()).To(Equal(newContent))
				Expect(urlNodeConcrete.GetUpdatedAt()).To(BeTemporally(">", originalTime))
			})
		})

		Context("Scrapped status", func() {
			It("should default to false", func() {
				Expect(urlNodeConcrete.IsScrapped()).To(BeFalse())
			})

			It("should set and get scrapped status", func() {
				originalTime := urlNodeConcrete.GetUpdatedAt()
				
				time.Sleep(time.Millisecond)
				urlNodeConcrete.SetScrapped(true)
				
				Expect(urlNodeConcrete.IsScrapped()).To(BeTrue())
				Expect(urlNodeConcrete.GetUpdatedAt()).To(BeTemporally(">", originalTime))
			})

			It("should update timestamp when changing scrapped status", func() {
				originalTime := urlNodeConcrete.GetUpdatedAt()
				
				time.Sleep(time.Millisecond)
				urlNodeConcrete.SetScrapped(false)
				
				Expect(urlNodeConcrete.IsScrapped()).To(BeFalse())
				Expect(urlNodeConcrete.GetUpdatedAt()).To(BeTemporally(">", originalTime))
			})
		})

		Context("Timestamps", func() {
			It("should have creation and update timestamps", func() {
				createdAt := urlNodeConcrete.GetCreatedAt()
				updatedAt := urlNodeConcrete.GetUpdatedAt()
				
				Expect(createdAt).NotTo(BeZero())
				Expect(updatedAt).NotTo(BeZero())
				Expect(updatedAt).To(BeTemporally(">=", createdAt))
			})

			It("should update timestamp when setting properties", func() {
				originalTime := urlNodeConcrete.GetUpdatedAt()
				
				time.Sleep(time.Millisecond)
				urlNode.SetProperty("test", "value")
				
				Expect(urlNodeConcrete.GetUpdatedAt()).To(BeTemporally(">", originalTime))
			})
		})
	})

	Describe("Type identification", func() {
		BeforeEach(func() {
			urlNode = object.NewURLNode(mgdsId, name, desc, url)
		})

		It("should return correct node type", func() {
			Expect(urlNode.GetType()).To(Equal("url"))
			Expect(urlNode.GetType()).To(Equal(object.URLNodeType))
		})
	})
})