package analyze_webpage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAnalyzeWebpage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AnalyzeWebpage Suite")
}
