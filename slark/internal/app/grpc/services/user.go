package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dchest/captcha"
	rd "github.com/go-redis/redis/v8"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/model"
	slark_mongo "github.com/nextsurfer/slark/internal/pkg/mongo"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"github.com/nextsurfer/slark/internal/pkg/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// UserService : service is pure business
type UserService struct {
	env         gutil.APPEnvType
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
	MongoDB     *mongo.Database
}

// NewUserService is factory function
func NewUserService(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, mongoDB *mongo.Database) *UserService {
	return &UserService{
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
		MongoDB:     mongoDB,
	}
}
func (s *UserService) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

// LoginByPhone api
func (s *UserService) LoginByPhone(ctx context.Context, rpcCtx *rpc.Context, phone, passwordHash string) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	user, err := s.daoManager.UserDAO.GetFromPhone(ctx, phone)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(gerror.ErrorCodeMysqlRead)
	}
	if user == nil {
		err = fmt.Errorf("can't find account by phone number : %s", phone)
		rpcCtx.Logger.Error("application error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_PhoneNotRegistered")).WithCode(response.StatusCodeLoginNotRegistered)
	}
	if user.PasswordHash != passwordHash {
		err = fmt.Errorf("user input error password for the account with phone number : %s", phone)
		rpcCtx.Logger.Error("application error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongPassword")).WithCode(response.StatusCodeInvalidPassword)
	}
	return s.loginWithUser(ctx, rpcCtx, user)
}

// LoginByEmail api
func (s *UserService) LoginByEmail(ctx context.Context, rpcCtx *rpc.Context, email, passwordHash string) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// 1. check mysql
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("email [%s] has not registered", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_EmailNotRegistered")).
			WithCode(response.StatusCodeRegisterEmailNotExists)
	}
	if user.PasswordHash != passwordHash {
		err = fmt.Errorf("user input error password for the account with email : %s", email)
		rpcCtx.Logger.Error("application error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongPassword")).WithCode(response.StatusCodeInvalidPassword)
	}
	// 2. log in with user
	return s.loginWithUser(ctx, rpcCtx, user)
}

// LoginBySession login by session
func (s *UserService) LoginBySession(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// 1. check request has session id
	if rpcCtx.SessionID == "" {
		err := errors.New("login without session")
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}

	// 2. query redis with session id
	rdsClient := s.redisOption.Client

	// 3. if session id is expired, clean session id on client, return error
	session, err := util.GetSessionInRedis(ctx, s.redisOption.Client, rpcCtx.SessionID)
	if err == rd.Nil {
		// remove session in cookie
		util.RemoveSessionInCookie(ctx)
		err = fmt.Errorf("session id [%s] expired", rpcCtx.SessionID)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeLoginSessionExpired)
	} else if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(gerror.ErrorCodeRedisRead)
	}

	// 4. if session id is ok, update userInfo
	session.LoginIP = rpcCtx.IP
	session.SetSessionID(rpcCtx.SessionID)
	// 5. store session info in redis
	if err := util.UpdateSessionInRedis(ctx, rdsClient, session); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(gerror.ErrorCodeRedisCreate)
	}

	// 6. update login ip in mysql
	if err := s.daoManager.SessionDAO.UpdateLoginIPInSession(ctx, rpcCtx.SessionID, rpcCtx.IP); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// 7. return user info
	user, err := s.daoManager.UserDAO.GetFromID(ctx, session.UserID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("cached session [id=%s] not stored in database", rpcCtx.SessionID)
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
	}

	data := &slark_api.LoginResponse_Data{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
	}

	// 6. refresh cookie
	util.SetSessionInCookie(ctx, rpcCtx.SessionID)

	return data, nil
}

func (s *UserService) SendRegistrationEmailCaptcha(ctx context.Context, rpcCtx *rpc.Context, email string) (string, *gerror.AppError) {
	// validate email parameter
	matched, err := regexp.MatchString(`[\w]+(\.[\w]+)*@[\w]+(\.[\w])+`, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	} else if !matched {
		err = fmt.Errorf("deformed email [%s]", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_EmailIsDeformed")).WithCode(response.StatusCodeDeformedEmail)
	}

	// check email
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		err = fmt.Errorf("email [%s] has registered", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_EmailHasRegistered")).WithCode(response.StatusCodeRegisterEmailExists)
	}

	// generate captch code and saved to redis
	captcha, err := util.GenerateDigitCaptchaWithStoreFuncs(ctx, func(captcha string) error {
		return util.StoreRegistrationCodeInRedis(ctx, rpcCtx, s.redisOption.Client, email, captcha)
	})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// send captcha email
	if !s.isEnvTest() {
		if err := util.SendCaptchaEmail(ctx, email, captcha); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		// 将 验证码 存在mongodb上 （一个列表， 维持最多20条记录， 也就是要手动清理旧的数据）
		coll := s.MongoDB.Collection(slark_mongo.CollectionName_RegistrationCaptcha)
		// 1. get
		var registrationCaptcha slark_mongo.RegistrationCaptcha
		if err := coll.FindOne(ctx, bson.D{}).Decode(&registrationCaptcha); err != nil {
			if err != mongo.ErrNoDocuments {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			} else {
				// new the record
				if _, err := coll.InsertOne(ctx, slark_mongo.RegistrationCaptcha{
					EmailCaptchas: []slark_mongo.EmailCaptcha{{Email: email, Captcha: captcha, CreatedAt: time.Now().UnixMilli()}},
				}); err != nil {
					rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
					return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
			}
		} else {
			// 2. update
			var list []slark_mongo.EmailCaptcha
			list = append(list, slark_mongo.EmailCaptcha{Email: email, Captcha: captcha, CreatedAt: time.Now().UnixMilli()})
			if len(registrationCaptcha.EmailCaptchas) < 20 {
				registrationCaptcha.EmailCaptchas = append(list, registrationCaptcha.EmailCaptchas...)
			} else {
				registrationCaptcha.EmailCaptchas = append(list, registrationCaptcha.EmailCaptchas[:19]...)
			}
			if _, err := coll.UpdateByID(ctx, registrationCaptcha.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "emailCaptchas", Value: registrationCaptcha.EmailCaptchas}}}}); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	return captcha, nil
}

// RegisterByEmail Register By Email
func (s *UserService) RegisterByEmail(ctx context.Context, rpcCtx *rpc.Context, email, nickname, password, captcha string) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// check nickname in db
	user, err := s.daoManager.UserDAO.GetByNickname(ctx, nickname)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		// it's a repeated nickname
		err = fmt.Errorf("nickname [%s] has registered", nickname)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("RegisterErrMsg_NicknameHasRegistered")).WithCode(response.StatusCodeRepeatedNickname)
	}

	// validate captcha
	if captcha != "000000" || !s.isEnvTest() {
		if err := util.ValidateRegistrationCodeInRedis(ctx, rpcCtx, s.redisOption.Client, email, captcha); err != nil {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_PasswordIsRequired")).WithCode(response.StatusCodeBadRequest)
		}
	}

	// validate register info to avoid duplicate registration
	user, err = s.daoManager.UserDAO.GetFromEmail(ctx, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		err = fmt.Errorf("email [%s] has registered", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("RegisterErrMsg_EmailHasRegistered")).WithCode(response.StatusCodeRegisterEmailExists)
	}

	// 3. write to mysql
	account := &model.SlkUser{
		Nickname:     nickname,
		PasswordHash: password,
		Email:        email,
	}
	if err := s.daoManager.UserDAO.Create(ctx, account); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 4. login and session
	return s.loginWithUser(ctx, rpcCtx, account)
}

func (s *UserService) RegisterByEmailCaptcha(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.RegisterByEmailCaptchaRequest) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// 1. validate captcha
	if req.Captcha != "000000" || !s.isEnvTest() {
		if err := util.ValidateRegistrationCodeInRedis(ctx, rpcCtx, s.redisOption.Client, req.Email, req.Captcha); err != nil {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_PasswordIsRequired")).WithCode(response.StatusCodeBadRequest)
		}
	}
	// 2. validate register info to avoid duplicate registration
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, req.Email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		err = fmt.Errorf("email [%s] has registered", req.Email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("RegisterErrMsg_EmailHasRegistered")).WithCode(response.StatusCodeRegisterEmailExists)
	}
	// 3. create a record
	account := &model.SlkUser{
		Email: req.Email,
	}
	if err := s.daoManager.UserDAO.Create(ctx, account); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 4. update nickname
	nickname := fmt.Sprintf("user%d", account.ID)
	if err := s.daoManager.UserDAO.UpdateNickname(ctx, account.ID, nickname); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	account.Nickname = nickname
	// 5. login and session
	return s.loginWithUser(ctx, rpcCtx, account)
}

// LoginByApple login or register
func (s *UserService) LoginByApple(ctx context.Context, rpcCtx *rpc.Context, email, userIdentifier string) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	if email == "" {
		// if login
		thirdPartyInfo, err := s.daoManager.ThirdPartyDAO.GetFromOpenID(ctx, util.ApplicationNameForContext(rpcCtx), userIdentifier, util.ThirdPartyApple)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if thirdPartyInfo == nil {
			err = errors.New("login by apple fatal error : maybe last registeration failed ; maybe some data lost in backend")
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
		}
		user, err := s.daoManager.UserDAO.GetFromID(ctx, thirdPartyInfo.UserID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if user == nil {
			err = fmt.Errorf("login by apple fatal error : cannot get user info from slk_user by id %d", thirdPartyInfo.UserID)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
		}
		return s.loginWithUser(ctx, rpcCtx, user)
	}

	// check third party account again to avoid some error in client
	thirdPartyInfo, err := s.daoManager.ThirdPartyDAO.GetFromOpenID(ctx, util.ApplicationNameForContext(rpcCtx), userIdentifier, util.ThirdPartyApple)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if thirdPartyInfo != nil {
		err = errors.New("some error happend in client. third party info has been written to mysql but client try to register again")
		rpcCtx.Logger.Warn("bad request", zap.NamedError("appError", err))
		// goto login
		user, err := s.daoManager.UserDAO.GetFromID(ctx, thirdPartyInfo.UserID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if user == nil {
			err = fmt.Errorf("login by apple fatal error : cannot get user info from slk_user by id %d", thirdPartyInfo.UserID)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
		}
		return s.loginWithUser(ctx, rpcCtx, user)
	}

	// check email
	matched, err := regexp.MatchString(`[\w]+(\.[\w]+)*@[\w]+(\.[\w])+`, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	} else if !matched {
		err = fmt.Errorf("deformed email [%s]", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_EmailIsDeformed")).WithCode(response.StatusCodeDeformedEmail)
	}
	// check nickname
	nickname := strings.Split(email, "@")[0]
	user, err := s.daoManager.UserDAO.GetByNickname(ctx, nickname)
	if err != nil {
		rpcCtx.Logger.Error("internel error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongNickname")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		// this is a repeated nickname
		randomDigits := captcha.RandomDigits(6)
		nickname = fmt.Sprintf("%s%d%d%d%d%d%d",
			nickname,
			randomDigits[0],
			randomDigits[1],
			randomDigits[2],
			randomDigits[3],
			randomDigits[4],
			randomDigits[5],
		)
	}

	// if register, write to mysql.
	//     use transaction
	tx, txDAO := s.daoManager.Transaction()
	account := &model.SlkUser{
		Nickname: nickname,
		Email:    email,
	}
	if err := txDAO.UserDAO.Create(ctx, account); err != nil {
		if err := tx.Rollback().Error; err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		}
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	thirdPartyInfo = &model.SlkThirdParty{
		Application: util.ApplicationNameForContext(rpcCtx),
		UserID:      account.ID,
		OpenID:      userIdentifier,
		ThirdParty:  util.ThirdPartyApple,
	}
	if err := txDAO.ThirdPartyDAO.Create(ctx, thirdPartyInfo); err != nil {
		if err := tx.Rollback().Error; err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		}
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err := tx.Commit().Error; err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// then login
	return s.loginWithUser(ctx, rpcCtx, account)
}

// LogOutBySession login out.
func (s *UserService) LogOutBySession(ctx context.Context, rpcCtx *rpc.Context) *gerror.AppError {
	// 1. check request has session id
	if rpcCtx.SessionID == "" {
		err := errors.New("logout without session id")
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}

	session, err := s.daoManager.SessionDAO.GetFromSessionID(ctx, rpcCtx.SessionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if session == nil {
		// query redis with session id
		util.DeleteSessionInRedis(ctx, rpcCtx, s.redisOption.Client)
		// remove session in cookie
		util.RemoveSessionInCookie(ctx)
		return nil
	}

	// 2. delete session in mysql.
	if err := s.daoManager.SessionDAO.DeleteBySessionID(ctx, session.SessionID); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// 3. query redis with session id
	if err := util.DeleteSessionInRedis(ctx, rpcCtx, s.redisOption.Client); err != nil {
		// TODO redis事务
		rpcCtx.Logger.Error("LogOutBySession delete session in redis failed", zap.NamedError("appError", err))
	}

	// 4.remove session in cookie
	util.RemoveSessionInCookie(ctx)
	return nil
}

func (s *UserService) CheckLogin(ctx context.Context, rpcCtx *rpc.Context) *gerror.AppError {
	if rpcCtx.SessionID == "" {
		err := errors.New("login without session")
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}

	// get session info from cache
	sessionInfo, err := util.GetSessionInRedis(ctx, s.redisOption.Client, rpcCtx.SessionID)
	if err != nil {
		if err != rd.Nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	if sessionInfo != nil {
		if sessionInfo.DeviceID != rpcCtx.DeviceID {
			err = fmt.Errorf("user with session[%s] not login on device[%s]", rpcCtx.SessionID, rpcCtx.DeviceID)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeUnauthorized)
		}
		return nil
	}

	// get session from db if no cache
	session, err := s.daoManager.SessionDAO.GetFromSessionID(ctx, rpcCtx.SessionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if session == nil {
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeUnauthorized)
	}
	// cache the session
	sessionInfo = &util.SessionInfo{
		LoginIP:  session.LoginIP,
		UserID:   session.UserID,
		DeviceID: session.DeviceID,
	}
	sessionInfo.SetSessionID(rpcCtx.SessionID)
	_ = util.UpdateSessionInRedis(ctx, s.redisOption.Client, sessionInfo)

	if session.DeviceID != rpcCtx.DeviceID {
		err = fmt.Errorf("user with session[%s] not login on device[%s]", rpcCtx.SessionID, rpcCtx.DeviceID)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeUnauthorized)
	}
	return nil
}

func (s *UserService) handleLoginInfoWithSessionInfo(ctx context.Context, rpcCtx *rpc.Context, sessionInfo *util.SessionInfo) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// check device
	if sessionInfo.DeviceID != rpcCtx.DeviceID {
		err := fmt.Errorf("user with session[%s] not login on device[%s]", rpcCtx.SessionID, rpcCtx.DeviceID)
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeUnauthorized)
	}
	// 先从缓存获取登录状态信息，没有缓存则查库
	loginInfo, err := util.GetLoginInfoInRedis(ctx, s.redisOption.Client, rpcCtx.SessionID)
	if err != nil {
		if err != rd.Nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	if loginInfo != nil {
		return &slark_api.LoginResponse_Data{
			UserID:   loginInfo.UserID,
			Nickname: loginInfo.Nickname,
			Email:    loginInfo.Email,
			Phone:    loginInfo.Phone,
		}, nil
	}
	user, err := s.daoManager.UserDAO.GetFromID(ctx, sessionInfo.UserID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeUnauthorized)
	}
	// 将登录状态信息缓存
	if err := util.UpdateLoginInfoInRedis(ctx, s.redisOption.Client, &util.LoginInfo{
		SessionID: rpcCtx.SessionID,
		UserID:    user.ID,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &slark_api.LoginResponse_Data{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
	}, nil
}

func (s *UserService) LoginInfo(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	if rpcCtx.SessionID == "" {
		err := errors.New("login without session")
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}
	// get session info from cache
	sessionInfo, err := util.GetSessionInRedis(ctx, s.redisOption.Client, rpcCtx.SessionID)
	if err != nil {
		if err != rd.Nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	if sessionInfo != nil {
		return s.handleLoginInfoWithSessionInfo(ctx, rpcCtx, sessionInfo)
	}
	// get session from db if no cache
	session, err := s.daoManager.SessionDAO.GetFromSessionID(ctx, rpcCtx.SessionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if session == nil {
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeUnauthorized)
	}
	// cache the session
	sessionInfo = &util.SessionInfo{
		LoginIP:  session.LoginIP,
		UserID:   session.UserID,
		DeviceID: session.DeviceID,
	}
	sessionInfo.SetSessionID(rpcCtx.SessionID)
	_ = util.UpdateSessionInRedis(ctx, s.redisOption.Client, sessionInfo)
	return s.handleLoginInfoWithSessionInfo(ctx, rpcCtx, sessionInfo)
}

func (s *UserService) CheckRegistration(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.CheckRegistrationRequest) (*slark_api.CheckRegistrationResponse_Data, *gerror.AppError) {
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, req.Email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var data slark_api.CheckRegistrationResponse_Data
	if user != nil {
		data.Id = user.ID
		data.Nickname = user.Nickname
	}
	return &data, nil
}

func (s *UserService) ValidateUserIDs(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.ValidateUserIDsRequest) (*slark_api.ValidateUserIDsResponse_Data, *gerror.AppError) {
	users, err := s.daoManager.UserDAO.GetFromIDs(ctx, req.UserIDs)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []bool
	for _, userID := range req.UserIDs {
		var valid bool
		for _, user := range users {
			if user.ID == userID {
				valid = true
				break
			}
		}
		list = append(list, valid)
	}
	return &slark_api.ValidateUserIDsResponse_Data{List: list}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.GetUserInfoRequest) (*slark_api.GetUserInfoResponse_Data, *gerror.AppError) {
	user, err := s.daoManager.UserDAO.GetFromID(ctx, req.Id)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("cannot get user info by id %d", req.Id)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	return &slark_api.GetUserInfoResponse_Data{Id: user.ID, Nickname: user.Nickname, Email: user.Email}, nil
}

// log in and create session
func (s *UserService) loginWithUser(ctx context.Context, rpcCtx *rpc.Context, user *model.SlkUser) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// 1. delete old session in mysql
	if err := s.daoManager.SessionDAO.DeleteSession(ctx, user.ID, rpcCtx.DeviceID, util.ApplicationNameForContext(rpcCtx)); err != nil {
		rpcCtx.Logger.Error("delete session failed during login",
			zap.Int64("userID", user.ID),
			zap.String("application", "test"))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// 2. create and store new session
	sessionID := util.NewUUIDHexEncoding()
	sessionObj := &model.SlkSession{
		Application: util.ApplicationNameForContext(rpcCtx),
		UserID:      user.ID,
		SessionID:   sessionID,
		DeviceID:    rpcCtx.DeviceID,
		LoginIP:     rpcCtx.IP,
	}
	if err := s.daoManager.SessionDAO.Create(ctx, sessionObj); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// 3. update redis
	rdsClient := s.redisOption.Client

	session := util.SessionInfo{
		ExtraInfo: make(map[string]string),
		LoginIP:   rpcCtx.IP,
		UserID:    user.ID,
		DeviceID:  rpcCtx.DeviceID,
	}
	session.SetSessionID(sessionID)
	// 4. store session info in redis
	if err := util.UpdateSessionInRedis(ctx, rdsClient, &session); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// 5. model response
	data := &slark_api.LoginResponse_Data{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
	}

	// 6. set cookie
	util.SetSessionInCookie(ctx, sessionID)

	return data, nil
}

func (s *UserService) QRLogin(ctx context.Context, rpcCtx *rpc.Context, token string) *gerror.AppError {
	// validate token
	if err := util.QRLoginTokenExistsInRedis(ctx, rpcCtx, s.redisOption.Client, token); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_TokenIsInvalid")).WithCode(response.StatusCodeBadRequest)
	}

	// check login by session id
	if appErr := s.CheckLogin(ctx, rpcCtx); appErr != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", appErr.Error))
		return appErr
	}

	// update token cache for checking by web endpoint
	if err := util.StoreQRLoginTokenInRedis(ctx, rpcCtx, s.redisOption.Client, token); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) SendLoginEmailCode(ctx context.Context, rpcCtx *rpc.Context, email string) (string, *gerror.AppError) {
	// check email
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("email [%s] has not registered", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_EmailNotRegistered")).WithCode(response.StatusCodeRegisterEmailNotExists)
	}

	// generate captch code and saved to redis
	captcha, err := util.GenerateDigitCaptchaWithStoreFuncs(ctx, func(captcha string) error {
		return util.StoreLoginCodeInRedis(ctx, rpcCtx, s.redisOption.Client, email, captcha)
	})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_PasswordIsRequired")).WithCode(response.StatusCodeInternalServerError)
	}

	// send captcha email
	if !s.isEnvTest() {
		if err := util.SendCaptchaEmail(ctx, email, captcha); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		// 将 验证码 存在mongodb上 （一个列表， 维持最多20条记录， 也就是要手动清理旧的数据）
		coll := s.MongoDB.Collection(slark_mongo.CollectionName_LoginCaptcha)
		// 1. get
		var loginCaptcha slark_mongo.LoginCaptcha
		if err := coll.FindOne(ctx, bson.D{}).Decode(&loginCaptcha); err != nil {
			if err != mongo.ErrNoDocuments {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			} else {
				// new the record
				if _, err := coll.InsertOne(ctx, slark_mongo.LoginCaptcha{
					EmailCaptchas: []slark_mongo.EmailCaptcha{{Email: email, Captcha: captcha, CreatedAt: time.Now().UnixMilli()}},
				}); err != nil {
					rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
					return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
			}
		} else {
			// 2. update
			var list []slark_mongo.EmailCaptcha
			list = append(list, slark_mongo.EmailCaptcha{Email: email, Captcha: captcha, CreatedAt: time.Now().UnixMilli()})
			if len(loginCaptcha.EmailCaptchas) < 20 {
				loginCaptcha.EmailCaptchas = append(list, loginCaptcha.EmailCaptchas...)
			} else {
				loginCaptcha.EmailCaptchas = append(list, loginCaptcha.EmailCaptchas[:19]...)
			}
			if _, err := coll.UpdateByID(ctx, loginCaptcha.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "emailCaptchas", Value: loginCaptcha.EmailCaptchas}}}}); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return captcha, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}

	return captcha, nil
}

func (s *UserService) LoginByEmailCode(ctx context.Context, rpcCtx *rpc.Context, email, code string) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// check email
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("email [%s] has not registered", email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_EmailNotRegistered")).WithCode(response.StatusCodeRegisterEmailNotExists)
	}
	// validate captcha code
	if code == "000000" {
		if s.isEnvTest() {
			return s.loginWithUser(ctx, rpcCtx, user)
		}
	}
	if err := util.ValidateLoginCodeInRedis(ctx, rpcCtx, s.redisOption.Client, email, code); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_InvalidEmailCaptcha")).WithCode(response.StatusCodeBadRequest)
	}
	return s.loginWithUser(ctx, rpcCtx, user)
}

func (s *UserService) UpdateNickname(ctx context.Context, rpcCtx *rpc.Context, nickname string) *gerror.AppError {
	// login info
	loginInfo, gErr := s.LoginInfo(ctx, rpcCtx)
	if gErr != nil {
		return gErr
	}
	// check nickname
	user, err := s.daoManager.UserDAO.GetByNickname(ctx, nickname)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		if user.ID != loginInfo.UserID {
			err = fmt.Errorf("nickname [%s] is repeated", nickname)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_NicknameIsRepeated")).WithCode(response.StatusCodeRepeatedNickname)
		} else {
			return nil
		}
	}
	if err := s.daoManager.UserDAO.UpdateNickname(ctx, loginInfo.UserID, nickname); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) Unregister(ctx context.Context, rpcCtx *rpc.Context) *gerror.AppError {
	if rpcCtx.SessionID == "" {
		err := errors.New("login without session")
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}

	// check session and password
	session, err := s.daoManager.SessionDAO.GetFromSessionID(ctx, rpcCtx.SessionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if session == nil {
		err = fmt.Errorf("invalid session id [%s]: %v", rpcCtx.SessionID, err)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeBadRequest)
	}
	user, err := s.daoManager.UserDAO.GetFromID(ctx, session.UserID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("invalid user id [%d]: %v", session.UserID, err)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeBadRequest)
	}

	tx, daoManager := s.daoManager.Transaction()
	defer func() {
		if err != nil {
			rpcCtx.Logger.Error("transaction error", zap.NamedError("appError", err))
			if err := tx.Rollback().Error; err != nil {
				rpcCtx.Logger.Error("transaction rollback error", zap.NamedError("appError", err))
			}
		}
	}()

	// delete slk_session [cache] and slk_user
	if err = daoManager.UserDAO.DeleteByID(ctx, user.ID); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err = daoManager.SessionDAO.DeleteByID(ctx, session.ID); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err = daoManager.ThirdPartyDAO.DeleteByUserID(ctx, user.ID); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err := tx.Commit().Error; err != nil {
		rpcCtx.Logger.Error("transaction commit error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err = util.DeleteSessionInRedis(ctx, rpcCtx, s.redisOption.Client); err != nil {
		rpcCtx.Logger.Error("delete session cache failed", zap.NamedError("appError", err))
	}
	// delete cookie
	util.RemoveSessionInCookie(ctx)

	return nil
}

func (s *UserService) RandomNickname(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.RandomNicknameResponse_Data, *gerror.AppError) {
	rn, err := util.GenerateNickname("", 0, 0, "")
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// check the nickname
	user, err := s.daoManager.UserDAO.GetByNickname(ctx, rn)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user != nil {
		ri, err := util.GetRandomInt(0, 100000)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		rn = fmt.Sprintf("%s%d", rn, ri)
	}
	return &slark_api.RandomNicknameResponse_Data{Nickname: rn}, nil
}

func (s *UserService) CreateSecondaryPassword(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.CreateSecondaryPasswordRequest) *gerror.AppError {
	if rpcCtx.SessionID == "" {
		err := errors.New("login without session")
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}
	// check session and password
	session, err := s.daoManager.SessionDAO.GetFromSessionID(ctx, rpcCtx.SessionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if session == nil {
		err = fmt.Errorf("invalid session id [%s]: %v", rpcCtx.SessionID, err)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeBadRequest)
	}
	user, err := s.daoManager.UserDAO.GetFromID(ctx, session.UserID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("invalid user id [%d]: %v", session.UserID, err)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeBadRequest)
	}
	if user.SecondaryPasswordHash != "" {
		err = errors.New("the user's secondary password exists")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SecondaryPasswordExists")).WithCode(response.StatusCodeSecondaryPasswordExists)
	}
	// update user record
	if err := s.daoManager.UserDAO.UpdateSecondaryPassword(ctx, user.ID, req.PasswordHash); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) UpdateSecondaryPassword(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.UpdateSecondaryPasswordRequest) *gerror.AppError {
	if rpcCtx.SessionID == "" {
		err := errors.New("login without session")
		rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionEmpty")).WithCode(response.StatusCodeEmptySession)
	}
	// check session and password
	session, err := s.daoManager.SessionDAO.GetFromSessionID(ctx, rpcCtx.SessionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if session == nil {
		err = fmt.Errorf("invalid session id [%s]: %v", rpcCtx.SessionID, err)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeBadRequest)
	}
	user, err := s.daoManager.UserDAO.GetFromID(ctx, session.UserID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("invalid user id [%d]: %v", session.UserID, err)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_SessionExpired")).WithCode(response.StatusCodeBadRequest)
	}
	if user.SecondaryPasswordHash != req.OldPasswordHash {
		err = errors.New("the old password is wrong")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongPassword")).WithCode(response.StatusCodeUnauthorized)
	}
	// update user record
	if err := s.daoManager.UserDAO.UpdateSecondaryPassword(ctx, user.ID, req.NewPasswordHash); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) LoginBySecondaryPassword(ctx context.Context, rpcCtx *rpc.Context, req *slark_api.LoginBySecondaryPasswordRequest) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	// 1. check mysql
	user, err := s.daoManager.UserDAO.GetFromEmail(ctx, req.Email)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if user == nil {
		err = fmt.Errorf("email [%s] has not registered", req.Email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_EmailNotRegistered")).
			WithCode(response.StatusCodeRegisterEmailNotExists)
	}
	if user.SecondaryPasswordHash != req.PasswordHash {
		err = fmt.Errorf("user input wrong password for the account with email : %s", req.Email)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongPassword")).WithCode(response.StatusCodeInvalidPassword)
	}
	// 2. log in with user
	return s.loginWithUser(ctx, rpcCtx, user)
}
