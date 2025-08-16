package search_document_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSearchDocument(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SearchDocument Suite")
}