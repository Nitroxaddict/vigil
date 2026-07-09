package templates

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/Nitroxaddict/vigil/pkg/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var Funcs = template.FuncMap{
	"ToUpper":       strings.ToUpper,
	"ToLower":       strings.ToLower,
	"ToJSON":        toJSON,
	"Title":         cases.Title(language.AmericanEnglish).String,
	"versionOrID":   versionOrID,
	"sanitizeField": sanitizeField,
}

func toJSON(v interface{}) string {
	var bytes []byte
	var err error
	if bytes, err = json.MarshalIndent(v, "", "  "); err != nil {
		return fmt.Sprintf("failed to marshal JSON in notification template: %v", err)
	}
	return string(bytes)
}

// versionOrID returns the OCI version label from the latest image when present,
// otherwise the 12-char short ID of the latest image. Safe for both text and
// html/template engines: the returned string is auto-escaped by html/template.
func versionOrID(cr types.ContainerReport) string {
	if v := cr.LatestImageVersion(); v != "" {
		return v
	}
	return string(cr.LatestImageID().ShortID())
}

// sanitizeField strips ASCII control characters (< 0x20 and 0x7f) from s.
// Used in the email-text template to prevent CR/LF from breaking the
// line-oriented plain-text format.
func sanitizeField(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}
