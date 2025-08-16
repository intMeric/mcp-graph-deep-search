package keyword_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKeyword(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keyword Suite")
}
