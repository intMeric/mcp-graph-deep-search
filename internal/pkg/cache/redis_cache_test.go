package cache_test

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redismock/v9"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

var _ = Describe("Redis Cache", func() {
	var (
		redisCache *MockRedisCache
		mockClient redismock.ClientMock
		ctx        context.Context
	)

	BeforeEach(func() {
		var client *redis.Client
		client, mockClient = redismock.NewClientMock()

		redisCache = &MockRedisCache{client: client}
		ctx = context.Background()
	})

	AfterEach(func() {
		Expect(mockClient.ExpectationsWereMet()).To(Succeed())
	})

	Describe("Set", func() {
		Context("with successful Redis operations", func() {
			It("should store string values", func() {
				key := "test_string"
				value := "hello world"
				// The mock should expect the JSON-marshaled bytes
				expectedData := []byte(`"hello world"`)

				mockClient.ExpectSet(key, expectedData, 0).SetVal("OK")

				err := redisCache.Set(ctx, key, value, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should store values with expiration", func() {
				key := "test_expiry"
				value := "will expire"
				expiration := 5 * time.Minute
				expectedData := []byte(`"will expire"`)

				mockClient.ExpectSet(key, expectedData, expiration).SetVal("OK")

				err := redisCache.Set(ctx, key, value, expiration)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should store complex struct values", func() {
				type TestStruct struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}

				key := "test_struct"
				value := TestStruct{Name: "John", Age: 30}
				expectedData := []byte(`{"name":"John","age":30}`)

				mockClient.ExpectSet(key, expectedData, 0).SetVal("OK")

				err := redisCache.Set(ctx, key, value, 0)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with Redis errors", func() {
			It("should return error when Redis SET fails", func() {
				key := "test_error"
				value := "test"
				expectedData := []byte(`"test"`)

				mockClient.ExpectSet(key, expectedData, 0).SetErr(redis.ErrClosed)

				err := redisCache.Set(ctx, key, value, 0)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(redis.ErrClosed))
			})
		})
	})

	Describe("Get", func() {
		Context("with successful Redis operations", func() {
			It("should retrieve string values", func() {
				key := "test_string"
				expectedValue := "hello world"
				storedData := `"hello world"`

				mockClient.ExpectGet(key).SetVal(storedData)

				var result string
				err := redisCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expectedValue))
			})

			It("should retrieve struct values", func() {
				type TestStruct struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}

				key := "test_struct"
				expectedValue := TestStruct{Name: "John", Age: 30}
				storedData := `{"name":"John","age":30}`

				mockClient.ExpectGet(key).SetVal(storedData)

				var result TestStruct
				err := redisCache.Get(ctx, key, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expectedValue))
			})
		})

		Context("with Redis errors", func() {
			It("should return error when key doesn't exist", func() {
				key := "non_existent"

				mockClient.ExpectGet(key).RedisNil()

				var result string
				err := redisCache.Get(ctx, key, &result)
				Expect(err).To(HaveOccurred())
			})

			It("should return error when Redis GET fails", func() {
				key := "test_error"

				mockClient.ExpectGet(key).SetErr(redis.ErrClosed)

				var result string
				err := redisCache.Get(ctx, key, &result)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(redis.ErrClosed))
			})
		})
	})

	Describe("Delete", func() {
		Context("with successful Redis operations", func() {
			It("should delete existing keys", func() {
				key := "test_delete"

				mockClient.ExpectDel(key).SetVal(1)

				err := redisCache.Delete(ctx, key)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with Redis errors", func() {
			It("should return error when Redis DEL fails", func() {
				key := "test_error"

				mockClient.ExpectDel(key).SetErr(redis.ErrClosed)

				err := redisCache.Delete(ctx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(redis.ErrClosed))
			})
		})
	})

	Describe("Exists", func() {
		Context("with successful Redis operations", func() {
			It("should return true when key exists", func() {
				key := "test_exists"

				mockClient.ExpectExists(key).SetVal(1)

				exists, err := redisCache.Exists(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})

			It("should return false when key doesn't exist", func() {
				key := "test_not_exists"

				mockClient.ExpectExists(key).SetVal(0)

				exists, err := redisCache.Exists(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})

		Context("with Redis errors", func() {
			It("should return error when Redis EXISTS fails", func() {
				key := "test_error"

				mockClient.ExpectExists(key).SetErr(redis.ErrClosed)

				exists, err := redisCache.Exists(ctx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(redis.ErrClosed))
				Expect(exists).To(BeFalse())
			})
		})
	})
})

// MockRedisCache implements the same logic as RedisCache but with dependency injection for testing
type MockRedisCache struct {
	client *redis.Client
}

func (m *MockRedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return m.client.Set(ctx, key, data, expiration).Err()
}

func (m *MockRedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := m.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (m *MockRedisCache) Delete(ctx context.Context, key string) error {
	return m.client.Del(ctx, key).Err()
}

func (m *MockRedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
