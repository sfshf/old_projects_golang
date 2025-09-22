package simpleemail

import (
	"bytes"
	"context"
	"html/template"

	"gopkg.in/gomail.v2"
)

var (
	_emailServerHost                     = "box.n1xt.net"
	_emailServerPort                     = 465
	_emailUsername                       = "noreply@n1xt.net"
	_emailPassword                       = "xernyh-hyktyg13"
	_emailFrom                           = "noreply@n1xt.net"
	_cronJobNotificationHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Oracle Cron Notification</title>
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
				<h1>Oracle网关定时任务通知</h1>
				<br />
				<p>{{ .Msg }}</p>
				<br />
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func SendCronJobNotificationEmail(ctx context.Context, email, message string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", email)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "Oracle Cron Job Notification")
	// generate html text, and setted to email body
	data := struct {
		Msg string
	}{
		Msg: message,
	}
	var buf bytes.Buffer
	cronJobNotificationHtmlTemplate, err := template.New("CronJobNotificationHtmlTemplate").Parse(_cronJobNotificationHtmlTemplateText)
	if err != nil {
		return err
	}
	cronJobNotificationHtmlTemplate.Execute(&buf, data)
	msg.SetBody("text/html", buf.String())
	dialer := gomail.NewDialer(_emailServerHost, _emailServerPort, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
