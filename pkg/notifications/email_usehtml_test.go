package notifications_test

// Tests that UseHTML in the SMTP shoutrrr URL derives from the resolved
// notification-template flag value, not the plaintext flag independently.

import (
	"github.com/Nitroxaddict/vigil/cmd"
	"github.com/Nitroxaddict/vigil/internal/flags"
	"github.com/Nitroxaddict/vigil/pkg/notifications"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("email notifier UseHTML flag wiring", func() {
	// baseArgs are the minimum SMTP flags needed to build a valid URL.
	baseArgs := func(extra ...string) []string {
		args := []string{
			"--notifications", "email",
			"--notification-email-from", "from@example.com",
			"--notification-email-to", "to@example.com",
			"--notification-email-server", "mail.example.com",
			"--notification-email-server-user", "user",
			"--notification-email-server-password", "pass",
		}
		return append(args, extra...)
	}

	buildEmailURL := func(args []string) string {
		command := cmd.NewRootCommand()
		flags.RegisterNotificationFlags(command)
		Expect(command.ParseFlags(args)).To(Succeed())
		urls, _ := notifications.AppendLegacyUrls([]string{}, command)
		Expect(urls).To(HaveLen(1))
		return urls[0]
	}

	When("notification-template is explicitly set to email-html", func() {
		It("sets UseHTML=Yes in the SMTP URL", func() {
			args := baseArgs("--notification-template", "email-html")
			url := buildEmailURL(args)
			Expect(url).To(ContainSubstring("usehtml=Yes"),
				"UseHTML must be true when template is email-html")
		})
	})

	When("notification-template is explicitly set to email-text", func() {
		It("does not set UseHTML=Yes in the SMTP URL", func() {
			args := baseArgs("--notification-template", "email-text")
			url := buildEmailURL(args)
			Expect(url).NotTo(ContainSubstring("usehtml=Yes"),
				"UseHTML must be false when template is email-text")
		})
	})

	When("no notification-template is set (default)", func() {
		It("does not set UseHTML=Yes in the SMTP URL", func() {
			args := baseArgs()
			url := buildEmailURL(args)
			// Default template is not email-html, so UseHTML stays false.
			Expect(url).NotTo(ContainSubstring("usehtml=Yes"))
		})
	})
})
