package mailer

import (
	"bytes"
	"html/template"
)

type emailData struct {
	Body string
}

var emailTmpl = template.Must(template.New("email").Parse(`<!DOCTYPE html>
<html lang="cs">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;">
  <table role="presentation" cellpadding="0" cellspacing="0" width="100%" style="background-color:#f4f4f5;padding:48px 16px;">
    <tr>
      <td align="center">

        <table role="presentation" cellpadding="0" cellspacing="0" width="580" style="max-width:580px;width:100%;">

          <!-- Logo row -->
          <tr>
            <td style="padding:0 0 20px 0;">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 100" width="150" height="50" role="img" aria-label="Ticketa">
                <g transform="translate(4 18) scale(0.64)">
                  <rect x="3" y="3" width="94" height="94" rx="27" fill="#1B7A50"></rect>
                  <line x1="50" y1="14" x2="50" y2="86" stroke="#EDF7F1" stroke-width="3" stroke-dasharray="2.6 5" stroke-linecap="round" opacity="0.35"></line>
                  <path d="M30 33 H70 V44 H56 V70 H44 V44 H30 Z" fill="#ffffff"></path>
                </g>
                <text x="86" y="64" font-family="system-ui,-apple-system,'Segoe UI',sans-serif" font-size="48" font-weight="800" letter-spacing="-1.6" fill="#0F1714">Ticketa</text>
              </svg>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 4px rgba(0,0,0,0.08);">

              <!-- Accent bar -->
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
                <tr>
                  <td style="height:4px;background:#1B7A50;font-size:0;line-height:0;">&nbsp;</td>
                </tr>
              </table>

              <!-- Body -->
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
                <tr>
                  <td style="padding:40px 40px 36px;">
                    <p style="margin:0;font-size:16px;line-height:1.8;color:#18181b;">{{.Body}}</p>
                  </td>
                </tr>
              </table>

              <!-- Footer -->
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
                <tr>
                  <td style="padding:20px 40px 28px;border-top:1px solid #f1f1f1;">
                    <p style="margin:0;font-size:12px;color:#a1a1aa;line-height:1.6;">
                      Tato zpráva byla vygenerována automaticky systémem Ticketa.<br>
                      Prosím neodpovídejte na tento email.
                    </p>
                  </td>
                </tr>
              </table>

            </td>
          </tr>

          <!-- Bottom caption -->
          <tr>
            <td style="padding:20px 0 0 0;text-align:center;">
              <p style="margin:0;font-size:11px;color:#a1a1aa;">© Ticketa</p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

func renderHTML(body string) ([]byte, error) {
	var buf bytes.Buffer
	if err := emailTmpl.Execute(&buf, emailData{Body: body}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
