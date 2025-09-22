package util

import (
	"bytes"
	"context"
	"html/template"

	"gopkg.in/gomail.v2"
)

var (
	_emailServerHost              = "box.n1xt.net"
	_emailServerPort              = 465
	_emailUsername                = "noreply@n1xt.net"
	_emailPassword                = "xernyh-hyktyg13"
	_emailFrom                    = "noreply@n1xt.net"
	_captchaEmailHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Email Validation</title>
			<style type="text/css">
				:root {
					box-sizing: border-box;
				}
				*, ::before, ::after {
					box-sizing: inherit;
				}
				body {
					font-family: Arial, Helvetica, sans-serif;
					margin: 0;
				}
				.container {
					max-width: 680px;
					margin: 0 auto;
					text-align: center;
				}
			</style>
		</head>
		<body>
			<header>
				<div>
				</div>
			</header>
			<div class="container">
				<h1>验证您是该电子邮件地址的所有者</h1>
				<br />
				<p>{{ .Email }}</p>
				<br />
				<hr style="filter: alpha(opacity=100,finishopacity=0,style=3)" width="100%" color="#987cb9" size="3" />
				<br />
				<p>最近有人在验证电子邮件地址时输入了该电子邮件地址。</p>
				<br />
				<p>您可以使用此验证码来验证您是该电子邮件地址的所有者。</p>
				<br />
				<h3>{{ .CaptchaCode }}</h3>
				<br />
				<p>如果这不是您本人所为，则可能是有人误输了您的电子邮件地址。请勿将此验证码泄露给他人，您目前无需执行任何其他操作。</p>
				<br />
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func SendCaptchaEmail(ctx context.Context, email, captcha string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", email)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "Captcha Validation")
	// generate html text, and setted to email body
	data := struct {
		Email       string
		CaptchaCode string
	}{
		Email:       email,
		CaptchaCode: captcha,
	}
	var buf bytes.Buffer
	captchaEmailHtmlTemplate, err := template.New("CaptchaEmailHtmlTemplate").Parse(_captchaEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := captchaEmailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	dialer := gomail.NewDialer(_emailServerHost, _emailServerPort, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
