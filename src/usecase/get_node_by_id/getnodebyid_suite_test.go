package getnodebyid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetNodeById(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetNodeById Suite")
}