package spinnaker_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSpinnaker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spinnaker Suite")
}
