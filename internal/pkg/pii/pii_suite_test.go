package pii_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPII(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PII Suite")
}
