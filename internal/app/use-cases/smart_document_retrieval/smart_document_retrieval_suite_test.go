package smart_document_retrieval_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSmartDocumentRetrieval(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmartDocumentRetrieval Suite")
}