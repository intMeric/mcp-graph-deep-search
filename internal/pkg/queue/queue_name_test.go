package queue_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/queue"
)

var _ = Describe("Queue Name", func() {
	Describe("QueueName constants", func() {
		It("should have the expected queue names", func() {
			Expect(queue.Investigate).To(Equal(queue.QueueName("investigate")))
		})

		It("should be usable as string", func() {
			queueName := queue.Investigate
			Expect(string(queueName)).To(Equal("investigate"))
		})
	})
})
