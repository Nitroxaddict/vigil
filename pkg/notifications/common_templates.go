package notifications

var commonTemplates = map[string]string{
	`default-legacy`: "{{range .}}{{.Message}}{{println}}{{end}}",

	`default`: `
{{- if .Report -}}
  {{- with .Report -}}
    {{- if ( or .Updated .Failed ) -}}
{{len .Scanned}} Scanned, {{len .Updated}} Updated, {{len .Failed}} Failed
      {{- range .Updated}}
- {{.Name}} ({{.ImageName}}): {{.CurrentImageID.ShortID}} updated to {{.LatestImageID.ShortID}}
      {{- end -}}
      {{- range .Fresh}}
- {{.Name}} ({{.ImageName}}): {{.State}}
	  {{- end -}}
	  {{- range .Skipped}}
- {{.Name}} ({{.ImageName}}): {{.State}}: {{.Error}}
	  {{- end -}}
	  {{- range .Failed}}
- {{.Name}} ({{.ImageName}}): {{.State}}: {{.Error}}
	  {{- end -}}
    {{- end -}}
  {{- end -}}
{{- else -}}
  {{range .Entries -}}{{.Message}}{{"\n"}}{{- end -}}
{{- end -}}`,

	`porcelain.v1.summary-no-log`: `
{{- if .Report -}}
  {{- range .Report.All }}
    {{- .Name}} ({{.ImageName}}): {{.State -}}
    {{- with .Error}} Error: {{.}}{{end}}{{ println }}
  {{- else -}}
    no containers matched filter
  {{- end -}}
{{- end -}}`,

	`json.v1`: `{{ . | ToJSON }}`,

	// email-html is rendered via html/template (context-aware HTML escaping) to
	// prevent XSS from attacker-controlled container/image metadata. Do not
	// switch this to text/template.
	`email-html`: `{{- if .Report -}}
<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#e5e7eb;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
<div style="max-width:560px;margin:0 auto;padding:24px 16px 36px">
<div style="background:#0f172a;border-radius:8px 8px 0 0;padding:16px 20px 14px">
<div style="font-size:9px;font-weight:700;letter-spacing:2px;text-transform:uppercase;color:#475569;margin-bottom:6px">VIGIL</div>
<div style="display:flex;align-items:center;justify-content:space-between">
<span style="font-size:18px;font-weight:700;color:#f1f5f9;letter-spacing:-0.3px">{{.Title}}</span>
{{- if .Report.Failed}}<span style="color:#fca5a5;background:#450a0a;font-size:11px;font-weight:600;padding:3px 10px;border-radius:20px">{{len .Report.Failed}} failed</span>{{- end}}
</div>
<div style="font-size:11px;color:#475569">{{.Time.Format "Monday, 2 January 2006 · 15:04 MST"}}</div>
</div>
<div style="height:16px"></div>
{{- if or .Report.Failed .Report.Updated .Report.Stale}}
{{- if .Report.Failed}}
<div style="font-size:9px;font-weight:700;letter-spacing:1.5px;text-transform:uppercase;color:#f87171;padding:0 2px 7px">FAILED</div>
<div style="background:#ffffff;border-radius:6px;overflow:hidden;margin-bottom:16px;box-shadow:0 1px 3px rgba(0,0,0,0.07)">
{{- range $i,$c := .Report.Failed}}
<div style="border-left:3px solid #ef4444;padding:10px 14px{{if $i}};border-top:1px solid #f3f4f6{{end}}">
<div style="display:flex;justify-content:space-between;align-items:baseline">
<span style="font-size:13px;font-weight:600;color:#111827">{{$c.Name}}</span>
<span style="font-family:monospace;font-size:10px;background:#f3f4f6;color:#6b7280;padding:1px 5px;border-radius:3px">{{$c.CurrentImageID.ShortID}}</span>
</div>
<div style="font-size:10px;color:#9ca3af;margin-top:1px">{{$c.ImageName}}</div>
{{- if $c.Error}}<div style="font-size:11px;color:#ef4444;background:#fef2f2;padding:5px 8px;border-radius:4px;margin-top:6px;font-family:monospace">{{$c.Error}}</div>{{end}}
</div>
{{- end}}
</div>
{{- end}}
{{- if .Report.Updated}}
<div style="font-size:9px;font-weight:700;letter-spacing:1.5px;text-transform:uppercase;color:#9ca3af;padding:0 2px 7px">UPDATED</div>
<div style="background:#ffffff;border-radius:6px;overflow:hidden;margin-bottom:16px;box-shadow:0 1px 3px rgba(0,0,0,0.07)">
{{- range $i,$c := .Report.Updated}}
<div style="border-left:3px solid #10b981;padding:9px 14px{{if $i}};border-top:1px solid #f3f4f6{{end}}">
<div style="display:flex;justify-content:space-between;align-items:baseline">
<span style="font-size:13px;font-weight:600;color:#111827">{{$c.Name}}</span>
<span style="font-family:monospace;font-size:10px"><span style="background:#f3f4f6;color:#6b7280;padding:1px 5px;border-radius:3px">{{$c.CurrentImageID.ShortID}}</span><span style="color:#d1d5db;margin:0 4px">&#8594;</span><span style="background:#ecfdf5;color:#059669;padding:1px 5px;border-radius:3px">{{versionOrID $c}}</span></span>
</div>
<div style="font-size:10px;color:#9ca3af;margin-top:1px">{{$c.ImageName}}</div>
</div>
{{- end}}
</div>
{{- end}}
{{- if .Report.Stale}}
<div style="display:flex;align-items:center;gap:8px;padding:0 2px 7px"><span style="font-size:9px;font-weight:700;letter-spacing:1.5px;text-transform:uppercase;color:#9ca3af">UPDATE AVAILABLE</span><span style="font-size:9px;font-weight:600;letter-spacing:0.5px;text-transform:uppercase;color:#6b7280;background:#e5e7eb;padding:1px 6px;border-radius:10px">monitor only</span></div>
<div style="background:#ffffff;border-radius:6px;overflow:hidden;margin-bottom:24px;box-shadow:0 1px 3px rgba(0,0,0,0.07)">
{{- range $i,$c := .Report.Stale}}
<div style="border-left:3px solid #e5e7eb;padding:8px 14px{{if $i}};border-top:1px solid #f3f4f6{{end}}">
<div style="display:flex;align-items:center;justify-content:space-between">
<span style="font-size:12px;color:#6b7280">{{$c.ImageName}}</span>
<span style="font-family:monospace;font-size:10px;color:#9ca3af;background:#f9fafb;padding:1px 5px;border-radius:3px">{{versionOrID $c}}</span>
</div>
</div>
{{- end}}
</div>
{{- end}}
{{- else}}
<div style="background:#ffffff;border-radius:6px;overflow:hidden;margin-bottom:24px;box-shadow:0 1px 3px rgba(0,0,0,0.07)">
<div style="border-left:3px solid #10b981;padding:12px 14px">
<div style="font-size:13px;font-weight:600;color:#111827">&#10003; Everything is up to date</div>
<div style="font-size:10px;color:#9ca3af;margin-top:2px">No updates, failures, or pending changes</div>
</div>
</div>
{{- end}}
<div style="text-align:center;font-size:10px;color:#9ca3af">Vigil &#183; github.com/Nitroxaddict/vigil</div>
</div>
</body>
</html>
{{- else -}}
{{range .Entries}}{{.Message}}
{{end -}}
{{- end -}}`,

	`email-text`: `{{- if .Report -}}
{{.Title}}
{{- if .Report.Failed}}

FAILED ({{len .Report.Failed}})
{{range .Report.Failed -}}
{{.Name | sanitizeField}} | {{.ImageName | sanitizeField}} | {{.Error | sanitizeField}}
{{end}}{{- end -}}
{{- if .Report.Updated}}

UPDATED ({{len .Report.Updated}})
{{range .Report.Updated -}}
{{.Name | sanitizeField}} | {{.ImageName | sanitizeField}} | {{.CurrentImageID.ShortID}} -> {{versionOrID .}}
{{end}}{{- end -}}
{{- if .Report.Stale}}

UPDATE AVAILABLE -- MONITOR ONLY ({{len .Report.Stale}})
{{range .Report.Stale -}}
{{.ImageName | sanitizeField}} | {{versionOrID .}}
{{end}}{{- end}}
{{- if not (or .Report.Failed .Report.Updated .Report.Stale)}}

Everything is up to date. Nothing to report.
{{- end}}
{{- else -}}
{{range .Entries}}{{.Message}}
{{end -}}
{{- end -}}`,
}
