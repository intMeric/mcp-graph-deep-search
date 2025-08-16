package serializer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTML(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTML Suite")
}
