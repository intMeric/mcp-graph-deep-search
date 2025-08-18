package get_document_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetDocument(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetDocument Suite")
}