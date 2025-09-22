package mongo

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionName_RegistrationCaptcha = "registration_captcha"
)

type RegistrationCaptcha struct {
	ID            bson.ObjectID  `bson:"_id,omitempty"`
	EmailCaptchas []EmailCaptcha `bson:"emailCaptchas,omitempty"`
}

type EmailCaptcha struct {
	Email     string `bson:"email,omitempty"`
	Captcha   string `bson:"captcha,omitempty"`
	CreatedAt int64  `bson:"createdAt,omitempty"`
}

const (
	CollectionName_LoginCaptcha = "login_captcha"
)

type LoginCaptcha struct {
	ID            bson.ObjectID  `bson:"_id,omitempty"`
	EmailCaptchas []EmailCaptcha `bson:"emailCaptchas,omitempty"`
}
