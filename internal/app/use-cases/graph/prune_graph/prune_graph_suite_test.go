package prune_graph_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPruneGraph(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PruneGraph Suite")
}