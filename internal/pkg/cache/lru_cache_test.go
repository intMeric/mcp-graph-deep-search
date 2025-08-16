package cache_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/cache"
)

var _ = Describe("LRU Cache", func() {
	var (
		lruCache cache.Cache
		ctx      context.Context
	)

	BeforeEach(func() {
		var err error
		lruCache, err = cache.NewLRUCache(10) // Size 10
		Expect(err).NotTo(HaveOccurred())
		Expect(lruCache).NotTo(BeNil())

		ctx = context.Background()
	})

	Describe("Set and Get", func() {
		Context("with simple string values", func() {
			It("should store and retrieve string values", func() {
				key := "test_string"
				value := "hello world"

				err := lruCache.Set(ctx, key, value, 0)
				Expect(err).NotTo(HaveOccurred())

				var result string
				err = lruCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))
			})

			It("should store and retrieve integer values", func() {
				key := "test_int"
				value := 42

				err := lruCache.Set(ctx, key, value, 0)
				Expect(err).NotTo(HaveOccurred())

				var result int
				err = lruCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))
			})

			It("should store and retrieve struct values", func() {
				type TestStruct struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}

				key := "test_struct"
				value := TestStruct{Name: "John", Age: 30}

				err := lruCache.Set(ctx, key, value, 0)
				Expect(err).NotTo(HaveOccurred())

				var result TestStruct
				err = lruCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))
			})
		})

		Context("with expiration", func() {
			It("should expire items after the specified duration", func() {
				key := "test_expiry"
				value := "will expire"

				err := lruCache.Set(ctx, key, value, 50*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())

				// Should exist immediately
				var result string
				err = lruCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))

				// Should exist before expiration
				exists, err := lruCache.Exists(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())

				// Wait for expiration
				time.Sleep(100 * time.Millisecond)

				// Should not exist after expiration
				err = lruCache.Get(ctx, key, &result)
				Expect(err).To(Equal(cache.ErrKeyNotFound))

				exists, err = lruCache.Exists(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})

			It("should not expire items with zero expiration", func() {
				key := "test_no_expiry"
				value := "will not expire"

				err := lruCache.Set(ctx, key, value, 0)
				Expect(err).NotTo(HaveOccurred())

				// Wait a bit
				time.Sleep(50 * time.Millisecond)

				var result string
				err = lruCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))
			})
		})

		Context("with non-existent keys", func() {
			It("should return ErrKeyNotFound for non-existent keys", func() {
				var result string
				err := lruCache.Get(ctx, "non_existent_key", &result)
				Expect(err).To(Equal(cache.ErrKeyNotFound))
			})
		})
	})

	Describe("Delete", func() {
		It("should delete existing keys", func() {
			key := "test_delete"
			value := "to be deleted"

			err := lruCache.Set(ctx, key, value, 0)
			Expect(err).NotTo(HaveOccurred())

			// Verify it exists
			exists, err := lruCache.Exists(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			// Delete it
			err = lruCache.Delete(ctx, key)
			Expect(err).NotTo(HaveOccurred())

			// Verify it's gone
			exists, err = lruCache.Exists(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			var result string
			err = lruCache.Get(ctx, key, &result)
			Expect(err).To(Equal(cache.ErrKeyNotFound))
		})

		It("should not error when deleting non-existent keys", func() {
			err := lruCache.Delete(ctx, "non_existent_key")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Exists", func() {
		It("should return true for existing keys", func() {
			key := "test_exists"
			value := "exists"

			err := lruCache.Set(ctx, key, value, 0)
			Expect(err).NotTo(HaveOccurred())

			exists, err := lruCache.Exists(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return false for non-existent keys", func() {
			exists, err := lruCache.Exists(ctx, "non_existent_key")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})
	})

	Describe("LRU behavior", func() {
		It("should evict least recently used items when cache is full", func() {
			smallCache, err := cache.NewLRUCache(2) // Very small cache
			Expect(err).NotTo(HaveOccurred())

			// Fill the cache
			err = smallCache.Set(ctx, "key1", "value1", 0)
			Expect(err).NotTo(HaveOccurred())

			err = smallCache.Set(ctx, "key2", "value2", 0)
			Expect(err).NotTo(HaveOccurred())

			// Both should exist
			exists, _ := smallCache.Exists(ctx, "key1")
			Expect(exists).To(BeTrue())
			exists, _ = smallCache.Exists(ctx, "key2")
			Expect(exists).To(BeTrue())

			// Add a third item, should evict the least recently used
			err = smallCache.Set(ctx, "key3", "value3", 0)
			Expect(err).NotTo(HaveOccurred())

			// key1 should be evicted (least recently used)
			exists, _ = smallCache.Exists(ctx, "key1")
			Expect(exists).To(BeFalse())

			// key2 and key3 should still exist
			exists, _ = smallCache.Exists(ctx, "key2")
			Expect(exists).To(BeTrue())
			exists, _ = smallCache.Exists(ctx, "key3")
			Expect(exists).To(BeTrue())
		})
	})
})
