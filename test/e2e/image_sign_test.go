package integration

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Podman image sign", func() {
	var origGNUPGHOME string

	BeforeEach(func() {
		SkipIfRemote("podman-remote image sign is not supported")
		tempGNUPGHOME := filepath.Join(podmanTest.TempDir, "tmpGPG")
		err := os.Mkdir(tempGNUPGHOME, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		origGNUPGHOME = os.Getenv("GNUPGHOME")
		err = os.Setenv("GNUPGHOME", tempGNUPGHOME)
		Expect(err).ToNot(HaveOccurred())

	})

	AfterEach(func() {
		// There's no way to run gpg without an agent, so, clean up
		// after every test. No need to check error status.
		cmd := exec.Command("gpgconf", "--kill", "gpg-agent")
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		_ = cmd.Run()

		os.Setenv("GNUPGHOME", origGNUPGHOME)
	})

	It("podman sign image", Serial, func() {
		cmd := exec.Command("gpg", "--import", "sign/secret-key.asc")
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())
		sigDir := filepath.Join(podmanTest.TempDir, "test-sign")
		err = os.MkdirAll(sigDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())
		session := podmanTest.Podman([]string{"image", "sign", "--directory", sigDir, "--sign-by", "foo@bar.com", "docker://library/alpine"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		_, err = os.Stat(filepath.Join(sigDir, "library"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("podman sign --all multi-arch image", Serial, func() {
		cmd := exec.Command("gpg", "--import", "sign/secret-key.asc")
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())
		sigDir := filepath.Join(podmanTest.TempDir, "test-sign-multi")
		err = os.MkdirAll(sigDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())
		session := podmanTest.Podman([]string{"image", "sign", "--all", "--directory", sigDir, "--sign-by", "foo@bar.com", "docker://library/alpine"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		fInfos, err := os.ReadDir(filepath.Join(sigDir, "library"))
		Expect(err).ToNot(HaveOccurred())
		Expect(len(fInfos)).To(BeNumerically(">", 1), "len(fInfos)")
	})
})
