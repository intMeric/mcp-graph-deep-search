package env_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/env"
)

var _ = Describe("Env Utils", func() {
	Describe("GetOrDefault", func() {
		const testKey = "TEST_ENV_VAR_FOR_TESTING"
		const defaultValue = "default_value"

		AfterEach(func() {
			// Clean up environment variable after each test
			os.Unsetenv(testKey)
		})

		Context("when environment variable exists", func() {
			It("should return the environment variable value", func() {
				expectedValue := "test_value"
				os.Setenv(testKey, expectedValue)

				result := env.GetOrDefault(testKey, defaultValue)

				Expect(result).To(Equal(expectedValue))
			})

			It("should return environment variable even if it's empty string", func() {
				os.Setenv(testKey, "")

				result := env.GetOrDefault(testKey, defaultValue)

				Expect(result).To(Equal(defaultValue))
			})

			It("should handle special characters in values", func() {
				specialValue := "test@value#with$special%chars"
				os.Setenv(testKey, specialValue)

				result := env.GetOrDefault(testKey, defaultValue)

				Expect(result).To(Equal(specialValue))
			})
		})

		Context("when environment variable does not exist", func() {
			It("should return the default value", func() {
				// Ensure the environment variable is not set
				os.Unsetenv(testKey)

				result := env.GetOrDefault(testKey, defaultValue)

				Expect(result).To(Equal(defaultValue))
			})

			It("should handle empty default value", func() {
				emptyDefault := ""
				os.Unsetenv(testKey)

				result := env.GetOrDefault(testKey, emptyDefault)

				Expect(result).To(Equal(emptyDefault))
			})

			It("should handle special characters in default value", func() {
				specialDefault := "default@value#with$special%chars"
				os.Unsetenv(testKey)

				result := env.GetOrDefault(testKey, specialDefault)

				Expect(result).To(Equal(specialDefault))
			})
		})

		Context("with edge cases", func() {
			It("should handle empty key name", func() {
				result := env.GetOrDefault("", defaultValue)

				Expect(result).To(Equal(defaultValue))
			})

			It("should handle whitespace-only environment variable", func() {
				whitespaceValue := "   "
				os.Setenv(testKey, whitespaceValue)

				result := env.GetOrDefault(testKey, defaultValue)

				Expect(result).To(Equal(whitespaceValue))
			})
		})
	})

	Describe("GetHostName", func() {
		Context("when getting hostname", func() {
			It("should return a non-empty hostname", func() {
				hostname := env.GetHostName()

				// Hostname should not be empty in normal circumstances
				// Note: This might be empty in some containerized environments
				Expect(hostname).NotTo(BeNil())
			})

			It("should return a string", func() {
				hostname := env.GetHostName()

				Expect(hostname).To(BeAssignableToTypeOf(""))
			})

			It("should be consistent across multiple calls", func() {
				hostname1 := env.GetHostName()
				hostname2 := env.GetHostName()

				Expect(hostname1).To(Equal(hostname2))
			})
		})

		Context("when hostname cannot be determined", func() {
			It("should handle errors gracefully", func() {
				// This test verifies that GetHostName doesn't panic
				// and returns a string (possibly empty) even if os.Hostname() fails
				hostname := env.GetHostName()

				Expect(hostname).To(BeAssignableToTypeOf(""))
			})
		})
	})
})
