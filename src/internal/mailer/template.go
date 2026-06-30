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
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
                <tr>
                  <td>
                    <span style="font-size:13px;font-weight:700;color:#6366f1;letter-spacing:4px;text-transform:uppercase;">TICKETA</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 4px rgba(0,0,0,0.08);">

              <!-- Accent bar -->
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
                <tr>
                  <td style="height:4px;background:#6366f1;font-size:0;line-height:0;">&nbsp;</td>
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
