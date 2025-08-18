package search_and_analyze_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSearchAndAnalyze(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SearchAndAnalyze Suite")
}
