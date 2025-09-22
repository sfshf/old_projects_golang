package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/nextsurfer/monitor/internal/common/simplehttp"
	monitor_mongo "github.com/nextsurfer/monitor/internal/mongo"
	"github.com/wxpusher/wxpusher-sdk-go"
	"github.com/wxpusher/wxpusher-sdk-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gopkg.in/gomail.v2"
)

const (
	RedisKey_MessageNotificationConfig = "Tester::MessageNotificationConfig"
)

type MessageNotificationConfig struct {
	UseTelegram  bool     `json:"useTelegram"`
	UseWxpusher  bool     `json:"useWxpusher"`
	WxpusherUIDs []string `json:"wxpusherUIDs"`
	UseEmail     bool     `json:"useEmail"`
	Emails       []string `json:"emails"`
}

func SendMessageByTelegram(message string) error {
	resp, err := simplehttp.PostJsonRequest(
		`https://api.telegram.org/bot1678806156:AAE8cWdlygrGCHWmHElQHNJ0ZjOv1IRQGeg/sendMessage`,
		map[string]string{
			"Accept":     "*/*",
			"Host":       "api.telegram.org",
			"User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15`,
		},
		struct {
			ChatID string `json:"chat_id"`
			Text   string `json:"text"`
		}{
			ChatID: "1417969737",
			Text:   message,
		},
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("!!! Telegram Request Error: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("!!! Telegram Request Error: %s", simplehttp.ErrResponseStatusCodeNotEqualTo200)
	}
	return nil
}

func SendMessageByWxpusher(content, summary string, uids []string) error {
	if len(uids) == 0 {
		return nil
	}
	if _, err := wxpusher.SendMessage(&model.Message{
		AppToken:    "AT_4MTVNvvjupPNKFSxcg6k5ezDxhh25rBa",
		Content:     content,
		Summary:     summary,
		ContentType: 2,
		UIds:        uids,
	}); err != nil {
		return err
	}
	return nil
}

var (
	_emailServerHost              = "box.n1xt.net"
	_emailServerPort              = 465
	_emailUsername                = "noreply@n1xt.net"
	_emailPassword                = "xernyh-hyktyg13"
	_emailFrom                    = "noreply@n1xt.net"
	_MessageEmailHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Retrieve PSWDS Unlock Password</title>
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
			</header>
			<div class="container">
				<p>{{.Message}}</p>
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func SendEmail(email, message string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", email)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "NextSurfer Message Email")
	// generate html text, and setted to email body
	data := struct {
		Message string
	}{
		Message: message,
	}
	var buf bytes.Buffer
	emailHtmlTemplate, err := template.New("MessageEmailHtmlTemplate").Parse(_MessageEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := emailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	dialer := gomail.NewDialer(_emailServerHost, _emailServerPort, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

func SendMessage(ctx context.Context, redisCli *redisv8.Client, mongoDB *mongo.Database, message string) error {
	var config MessageNotificationConfig
	// get config from redis
	configJson, err := redisCli.Get(ctx, RedisKey_MessageNotificationConfig).Result()
	if err != nil {
		if err != redisv8.Nil {
			return err
		} else {
			// get config from mongo
			coll := mongoDB.Collection(monitor_mongo.CollectionName_MessageNotificationConfig)
			var configFromDB monitor_mongo.MessageNotificationConfig
			if err := coll.FindOne(ctx, bson.D{}).Decode(&configFromDB); err != nil {
				return err
			}
			config.UseTelegram = configFromDB.UseTelegram
			config.UseWxpusher = configFromDB.UseWxpusher
			config.WxpusherUIDs = configFromDB.WxpusherUIDs
			config.UseEmail = configFromDB.UseEmail
			config.Emails = configFromDB.Emails
		}
	} else {
		if err := json.Unmarshal([]byte(configJson), &config); err != nil {
			return err
		}
	}
	log := monitor_mongo.MessageNotificationLog{
		Message:   message,
		CreatedAt: time.Now().Unix(),
	}
	if config.UseTelegram {
		log.Mode |= monitor_mongo.Ltelegram
		go func() {
			SendMessageByTelegram(message)
		}()
	}
	if config.UseWxpusher {
		log.Mode |= monitor_mongo.Lwxpusher
		go func() {
			SendMessageByWxpusher(message, "NextSurfer Message", config.WxpusherUIDs)
		}()
	}
	if config.UseEmail {
		log.Mode |= monitor_mongo.Lemail
		for _, email := range config.Emails {
			go func(email string) {
				SendEmail(email, message)
			}(email)
		}
	}
	// add log
	mongoDB.Collection(monitor_mongo.CollectionName_MessageNotificationLog).InsertOne(ctx, log)
	return nil
}
