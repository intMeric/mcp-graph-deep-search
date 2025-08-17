package graph_pruner_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGraphPruner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GraphPruner Suite")
}