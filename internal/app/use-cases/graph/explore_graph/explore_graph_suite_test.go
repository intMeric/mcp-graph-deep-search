package explore_graph_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExploreGraph(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExploreGraph Suite")
}