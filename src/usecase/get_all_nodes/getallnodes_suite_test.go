package getallnodes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetAllNodes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetAllNodes Suite")
}