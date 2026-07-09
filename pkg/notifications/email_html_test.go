package notifications

// White-box tests for the email-html and email-text templates.
// Tests live in the notifications package so they can call unexported helpers
// such as getShoutrrrHTMLTemplate and getShoutrrrTemplate.

import (
	"bytes"
	"strings"
	"time"

	"github.com/Nitroxaddict/vigil/internal/actions/mocks"
	s "github.com/Nitroxaddict/vigil/pkg/session"
	t "github.com/Nitroxaddict/vigil/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

// --- helpers ----------------------------------------------------------------

// mockReport wraps slices of ContainerReports so tests can build arbitrary
// report states without going through the progress machinery.
type mockReport struct {
	updated []t.ContainerReport
	failed  []t.ContainerReport
	stale   []t.ContainerReport
	fresh   []t.ContainerReport
	skipped []t.ContainerReport
	scanned []t.ContainerReport
}

func (r *mockReport) Updated() []t.ContainerReport { return r.updated }
func (r *mockReport) Failed() []t.ContainerReport  { return r.failed }
func (r *mockReport) Stale() []t.ContainerReport   { return r.stale }
func (r *mockReport) Fresh() []t.ContainerReport   { return r.fresh }
func (r *mockReport) Skipped() []t.ContainerReport { return r.skipped }
func (r *mockReport) Scanned() []t.ContainerReport { return r.scanned }
func (r *mockReport) All() []t.ContainerReport {
	all := make([]t.ContainerReport, 0)
	all = append(all, r.updated...)
	all = append(all, r.failed...)
	all = append(all, r.stale...)
	all = append(all, r.fresh...)
	all = append(all, r.skipped...)
	return all
}

// injectableStatus is a ContainerReport whose fields are fully controllable —
// used to inject attacker-controlled values into template rendering tests.
type injectableStatus struct {
	name           string
	imageName      string
	currentImageID t.ImageID
	latestImageID  t.ImageID
	latestVersion  string
	errStr         string
}

func (c *injectableStatus) ID() t.ContainerID           { return "id-test" }
func (c *injectableStatus) Name() string                 { return c.name }
func (c *injectableStatus) CurrentImageID() t.ImageID    { return c.currentImageID }
func (c *injectableStatus) LatestImageID() t.ImageID     { return c.latestImageID }
func (c *injectableStatus) LatestImageVersion() string   { return c.latestVersion }
func (c *injectableStatus) ImageName() string            { return c.imageName }
func (c *injectableStatus) Error() string                { return c.errStr }
func (c *injectableStatus) State() string                { return "Updated" }

func renderHTMLTemplate(data Data) (string, error) {
	htmlTpl, err := getShoutrrrHTMLTemplate()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := htmlTpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderTextTemplate(data Data) (string, error) {
	tpl, err := getShoutrrrTemplate("email-text", false)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func staticData() StaticData {
	return StaticData{
		Title: "Vigil updates on atlantis",
		Host:  "atlantis",
		Time:  time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC),
	}
}

// --- test suite -------------------------------------------------------------

var _ = Describe("email-html template", func() {

	Describe("HTML escaping (XSS prevention)", func() {
		// Security-critical: container names, image names, OCI version labels,
		// and error strings are attacker-influenceable. html/template must escape
		// them so that no attacker-controlled string executes as HTML.
		//
		// Correct html/template escaping:
		//   "<script>alert(1)</script>"  → "&lt;script&gt;alert(1)&lt;/script&gt;"
		//   "\">" + "<img…"              → "&#34;&gt;&lt;img…"
		//
		// We verify that raw opening angle brackets do NOT appear where
		// attacker-controlled content is rendered — the angle bracket is
		// precisely what makes a tag boundary.

		It("escapes script tag in container name (failed section)", func() {
			cs := &injectableStatus{
				name:           `<script>alert(1)</script>`,
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
				errStr:         "restart failed",
			}
			report := &mockReport{failed: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			// The raw <script tag must not appear — it must be escaped to &lt;script
			Expect(out).NotTo(ContainSubstring("<script>"),
				"raw <script> tag must not appear in HTML output")
			Expect(out).To(ContainSubstring("&lt;script&gt;"),
				"script tag must be HTML-escaped in output")
		})

		It("escapes img-onerror payload in container name (failed section)", func() {
			cs := &injectableStatus{
				name:           `"><img src=x onerror=alert(1)>`,
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
				errStr:         "restart failed",
			}
			report := &mockReport{failed: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			// The literal string "<img" must not appear as a raw tag start
			Expect(out).NotTo(ContainSubstring("<img "),
				"raw <img tag must not appear; must be escaped to &lt;img")
			Expect(out).To(ContainSubstring("&lt;img"),
				"img tag open must be escaped in output")
		})

		It("escapes XSS payloads in image name (updated section)", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      `<script>alert(1)</script>`,
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
			}
			report := &mockReport{updated: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("<script>"))
			Expect(out).To(ContainSubstring("&lt;script&gt;"))
		})

		It("escapes script tag in OCI version label (versionOrID, updated section)", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
				latestVersion:  `<script>alert(1)</script>`,
			}
			report := &mockReport{updated: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("<script>"),
				"script tag in OCI version label must be escaped")
			Expect(out).To(ContainSubstring("&lt;script&gt;"))
		})

		It("escapes img-onerror payload in OCI version label (versionOrID, updated section)", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
				latestVersion:  `"><img src=x onerror=alert(1)>`,
			}
			report := &mockReport{updated: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			// Raw <img must not appear in output
			Expect(out).NotTo(ContainSubstring("<img "),
				"raw <img tag in OCI version label must be escaped")
		})

		It("escapes XSS payload in error string (failed section)", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
				errStr:         `<script>alert(1)</script>`,
			}
			report := &mockReport{failed: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("<script>"))
			Expect(out).To(ContainSubstring("&lt;script&gt;"))
		})

		It("escapes XSS payloads in OCI version label (stale section)", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
				latestVersion:  `<script>alert(1)</script>`,
			}
			report := &mockReport{stale: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("<script>"))
		})
	})

	Describe("nil / missing label safety", func() {
		It("falls back to ShortID() when LatestImageVersion() is empty, no panic", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa000000000000"),
				latestImageID:  t.ImageID("sha256:bbbb000000000000"),
				latestVersion:  "", // no version label
			}
			report := &mockReport{updated: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			Expect(func() {
				out, err := renderHTMLTemplate(data)
				Expect(err).NotTo(HaveOccurred())
				// ShortID of sha256:bbbb000000000000 is first 12 hex chars after "sha256:"
				Expect(out).To(ContainSubstring("bbbb00000000"))
			}).NotTo(Panic())
		})

		It("falls back to ShortID() in email-text when LatestImageVersion() is empty", func() {
			cs := &injectableStatus{
				name:           "mycontainer",
				imageName:      "image:latest",
				currentImageID: t.ImageID("sha256:aaaa000000000000"),
				latestImageID:  t.ImageID("sha256:bbbb000000000000"),
				latestVersion:  "",
			}
			report := &mockReport{updated: []t.ContainerReport{cs}}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderTextTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("bbbb00000000"))
		})
	})

	Describe("'N failed' pill", func() {
		It("is absent when there are no failures", func() {
			report := mocks.CreateMockProgressReport(s.UpdatedState)
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("failed</span>"),
				"failed pill must not appear when no failures")
		})

		It("is present with the correct count when failures exist", func() {
			report := mocks.CreateMockProgressReport(s.FailedState, s.FailedState)
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("2 failed"), "pill must show the failure count")
		})
	})

	Describe("all-clear state", func() {
		It("renders a non-blank all-clear card in email-html when nothing in Failed/Updated/Stale", func() {
			// Use mockReport with empty failed/updated/stale to simulate a quiet run.
			report := &mockReport{}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeEmpty(), "all-clear email must not be empty")
			Expect(out).To(ContainSubstring("Everything is up to date"),
				"all-clear card text must appear")
		})

		It("renders a non-blank all-clear message in email-text when nothing in Failed/Updated/Stale", func() {
			report := &mockReport{}
			data := Data{StaticData: staticData(), Report: report}

			out, err := renderTextTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeEmpty())
			Expect(out).To(ContainSubstring("Everything is up to date"))
		})
	})

	Describe("nil report (startup / out-of-cycle log)", func() {
		// SendNotification(nil) is called on startup and for out-of-cycle logs.
		// The nil-guard in the template must prevent template execution errors
		// that would otherwise trigger the Fatal() path in sendEntries.

		It("email-html renders without error when report is nil", func() {
			entries := []*logrus.Entry{
				{Message: "startup log message"},
			}
			data := Data{StaticData: staticData(), Entries: entries, Report: nil}

			_, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
		})

		It("email-html renders empty string when both report and entries are nil", func() {
			data := Data{StaticData: staticData(), Entries: nil, Report: nil}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.TrimSpace(out)).To(BeEmpty(),
				"no report, no entries → nothing to send")
		})

		It("email-text renders without error when report is nil", func() {
			data := Data{StaticData: staticData(), Entries: []*logrus.Entry{{Message: "hello"}}, Report: nil}

			_, err := renderTextTemplate(data)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("mixed state rendering", func() {
		// Build a report directly with all three active buckets populated.
		buildMixedReport := func() *mockReport {
			failed := &injectableStatus{
				name: "fail1", imageName: "mock/fail1:latest",
				currentImageID: t.ImageID("sha256:ffff"),
				latestImageID:  t.ImageID("sha256:ffff"),
				errStr:         "restart failed",
			}
			updated := &injectableStatus{
				name: "updt1", imageName: "mock/updt1:latest",
				currentImageID: t.ImageID("sha256:aaaa"),
				latestImageID:  t.ImageID("sha256:bbbb"),
			}
			stale := &injectableStatus{
				name: "stale1", imageName: "mock/stale1:latest",
				currentImageID: t.ImageID("sha256:cccc"),
				latestImageID:  t.ImageID("sha256:dddd"),
			}
			return &mockReport{
				failed:  []t.ContainerReport{failed},
				updated: []t.ContainerReport{updated},
				stale:   []t.ContainerReport{stale},
			}
		}

		It("email-html renders failed, updated, and stale sections", func() {
			data := Data{StaticData: staticData(), Report: buildMixedReport()}

			out, err := renderHTMLTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("FAILED"))
			Expect(out).To(ContainSubstring("UPDATED"))
			Expect(out).To(ContainSubstring("UPDATE AVAILABLE"))
		})

		It("email-text renders all sections", func() {
			data := Data{StaticData: staticData(), Report: buildMixedReport()}

			out, err := renderTextTemplate(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("FAILED"))
			Expect(out).To(ContainSubstring("UPDATED"))
			Expect(out).To(ContainSubstring("UPDATE AVAILABLE -- MONITOR ONLY"))
		})
	})
})
