package queue_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/queue"
)

var _ = Describe("Queue Interface", func() {
	Describe("RequestMessage", func() {
		It("should have proper structure", func() {
			msg := queue.RequestMessage{
				ID:      "test-id",
				ReplyTo: "reply-queue",
				Data:    "test data",
			}

			Expect(msg.ID).To(Equal("test-id"))
			Expect(msg.ReplyTo).To(Equal("reply-queue"))
			Expect(msg.Data).To(Equal("test data"))
		})
	})

	Describe("ResponseMessage", func() {
		It("should have proper structure", func() {
			msg := queue.ResponseMessage{
				ID:    "test-id",
				Data:  "response data",
				Error: "",
			}

			Expect(msg.ID).To(Equal("test-id"))
			Expect(msg.Data).To(Equal("response data"))
			Expect(msg.Error).To(BeEmpty())
		})

		It("should handle error responses", func() {
			msg := queue.ResponseMessage{
				ID:    "test-id",
				Data:  nil,
				Error: "processing failed",
			}

			Expect(msg.ID).To(Equal("test-id"))
			Expect(msg.Data).To(BeNil())
			Expect(msg.Error).To(Equal("processing failed"))
		})
	})
})
