package webpage_analysis_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebpageAnalysis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WebpageAnalysis Suite")
}