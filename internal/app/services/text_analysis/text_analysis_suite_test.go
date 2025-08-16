package text_analysis_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTextAnalysis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Text Analysis Suite")
}
