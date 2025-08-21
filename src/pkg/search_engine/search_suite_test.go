package search_engine_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSearchEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SearchEngine Suite")
}