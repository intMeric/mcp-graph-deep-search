package graph_explorer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGraphExplorer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GraphExplorer Suite")
}