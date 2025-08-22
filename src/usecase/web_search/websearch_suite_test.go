package websearch_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebSearch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WebSearch UseCase Suite")
}
