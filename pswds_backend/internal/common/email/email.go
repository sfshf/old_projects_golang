package email

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

var (
	_emailServerHost = os.Getenv("EMAIL_SERVER_HOST")
	_emailServerPort = os.Getenv("EMAIL_SERVER_PORT")
	_emailUsername   = os.Getenv("EMAIL_SERVER_USERNAME")
	_emailPassword   = os.Getenv("EMAIL_SERVER_PASSWORD")
	_emailFrom       = os.Getenv("EMAIL_SERVER_FROM")
)

func SendEmail_RecoverUnlockPassword(ctx context.Context, email, ciphertext string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", email)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "Recover Unlock Password")
	// generate html text, and setted to email body
	data := struct {
		URL        string
		Ciphertext string
	}{
		URL:        "https://pwd.test.n1xt.net/decrypt",
		Ciphertext: ciphertext,
	}
	var buf bytes.Buffer
	captchaEmailHtmlTemplate, err := template.New("RecoverUnlockPasswordEmailHtmlTemplate").Parse(_recoverUnlockPasswordEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := captchaEmailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	port, err := strconv.Atoi(_emailServerPort)
	if err != nil {
		return err
	}
	dialer := gomail.NewDialer(_emailServerHost, port, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

const (
	_recoverUnlockPasswordEmailHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Recover PSWDS Unlock Password</title>
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
				<h1>请访问下述网站地址进行解锁密码找回。</h1>
				<br />
				<h3><a target="_blank" href="{{ .URL }}">{{ .URL }}</a></h3>
				<br />
				<hr style="filter: alpha(opacity=100,finishopacity=0,style=3)" width="100%" color="#987cb9" size="3" />
				<br />
				<p>访问网站地址成功后，请将下述密文复制到网站进行解锁密码找回操作</p>
				<br />
				<h3>{{ .Ciphertext }}</h3>
				<br />
				<p>如果这不是您本人所为，则可能是有人误输了您的电子邮件地址。请勿将此邮件内容泄露给他人，您目前无需执行任何其他操作。</p>
				<br />
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func SendEmail_TrustedContactBackupCiphertext(ctx context.Context, contactEmail, ciphertext string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", contactEmail)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "Recover Unlock Password")
	// generate html text, and setted to email body
	data := struct {
		URL        string
		Ciphertext string
	}{
		URL:        "https://pwd.test.n1xt.net/tcDecrypt",
		Ciphertext: ciphertext,
	}
	var buf bytes.Buffer
	captchaEmailHtmlTemplate, err := template.New("RecoverUnlockPasswordEmailHtmlTemplate").Parse(_recoverUnlockPasswordEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := captchaEmailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	port, err := strconv.Atoi(_emailServerPort)
	if err != nil {
		return err
	}
	dialer := gomail.NewDialer(_emailServerHost, port, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

const (
	_rejectFamilyRecoverEmailHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Recover PSWDS Unlock Password</title>
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
				<h1>{{ .Helper }}申请协助找回解锁密码</h1>
				<br />
				<p>如果您不同意此次申请，请点击<a target="_blank" href="{{ .URL }}">拒绝</a>，来拒绝此次找回申请。</p>
				<br />
				<p>为了保障您的数据安全，请勿将此邮件内容泄露给他人。</p>
				<br />
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func SendEmail_FamilyRecover(ctx context.Context, contactEmail, helper, recoverUUID string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", contactEmail)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "Recover Unlock Password")
	// generate html text, and setted to email body
	data := struct {
		Helper string
		URL    string
	}{
		Helper: helper,
		URL:    "https://pwd.test.n1xt.net/rejectFamilyRecover?uuid=" + recoverUUID,
	}
	var buf bytes.Buffer
	captchaEmailHtmlTemplate, err := template.New("RejectFamilyRecoverEmailHtmlTemplate").Parse(_rejectFamilyRecoverEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := captchaEmailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	port, err := strconv.Atoi(_emailServerPort)
	if err != nil {
		return err
	}
	dialer := gomail.NewDialer(_emailServerHost, port, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

const (
	_confirmFamilyRecoverEmailHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Recover PSWDS Unlock Password</title>
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
				<h1>主动找回解锁密码申请</h1>
				<br />
				<p>如果您同意此次申请，请点击<a target="_blank" href="{{ .URL }}">确认</a>，来同意此次找回申请。</p>
				<br />
				<p>为了保障您的数据安全，请勿将此邮件内容泄露给他人。</p>
				<br />
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func SendEmail_SelfRecover(ctx context.Context, contactEmail, recoverUUID string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", contactEmail)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "Recover Unlock Password")
	// generate html text, and setted to email body
	data := struct {
		URL string
	}{
		URL: "https://pwd.test.n1xt.net/confirmFamilyRecover?uuid=" + recoverUUID,
	}
	var buf bytes.Buffer
	captchaEmailHtmlTemplate, err := template.New("ConfirmFamilyRecoverEmailHtmlTemplate").Parse(_confirmFamilyRecoverEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := captchaEmailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	port, err := strconv.Atoi(_emailServerPort)
	if err != nil {
		return err
	}
	dialer := gomail.NewDialer(_emailServerHost, port, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
