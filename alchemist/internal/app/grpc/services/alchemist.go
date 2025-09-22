package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	alchemist_api "github.com/nextsurfer/alchemist/api"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/alchemist/pkg/consts"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	devicecheck "github.com/rinchsan/device-check-go/v2"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AlchemistService struct {
	app                                  string
	env                                  gutil.APPEnvType
	logger                               *zap.Logger
	daoManager                           *dao.Manager
	redisOption                          *redis.Option
	cron                                 *cron.Cron
	handleAppStoreNotificationEntryID    cron.EntryID
	handleAppStoreNotificationCronStatus struct {
		Started           bool
		StartedOrStopedAt time.Time
		ScheduleSpec      string
		LastExecError     string
	}
}

func NewAlchemistService(appID string, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) (*AlchemistService, error) {
	s := &AlchemistService{
		app:         appID,
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	cron := cron.New()
	s.cron = cron
	// appstore notification handle cron entry
	intervalStr := strings.TrimSpace(os.Getenv("NOTIFICATION_HANDLE_INTERVAL"))
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid NOTIFICATION_HANDLE_INTERVAL environment variable: %w", err)
	}
	spec := fmt.Sprintf("@every %ds", interval)
	handleAppStoreNotificationEntryID, err := cron.AddFunc(spec, func() {
		if err := handleAppStoreNotification(s.daoManager, logger); err != nil {
			s.handleAppStoreNotificationCronStatus.LastExecError = err.Error()
		}
	})
	if err != nil {
		return nil, err
	}
	s.handleAppStoreNotificationEntryID = handleAppStoreNotificationEntryID
	// start cron jobs
	s.cron.Start()
	s.handleAppStoreNotificationCronStatus.Started = true
	s.handleAppStoreNotificationCronStatus.StartedOrStopedAt = time.Now()
	s.handleAppStoreNotificationCronStatus.ScheduleSpec = spec
	return s, nil
}

func (s *AlchemistService) GetAppAccountToken(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*alchemist_api.GetAppAccountTokenResponse_Data, *gerror.AppError) {
	// get uuid by user id, if has
	accountToken, err := s.daoManager.SlarkUserDAO.GetByUserID(ctx, userID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if accountToken == nil {
		uuid := util.NewUUIDString()
		if err := s.daoManager.SlarkUserDAO.Create(ctx, &SlarkUser{
			AppAccountToken: uuid,
			UserID:          userID,
		}); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		return &alchemist_api.GetAppAccountTokenResponse_Data{AppAccountToken: uuid}, nil
	}
	return &alchemist_api.GetAppAccountTokenResponse_Data{AppAccountToken: accountToken.AppAccountToken}, nil
}

func joinReferralProgram(ctx context.Context, rpcCtx *rpc.Context, now time.Time, daoManager *dao.Manager, referralCode *ReferralCode, appID string, userID int64) error {
	if err := daoManager.ReferralCodeDAO.Create(ctx, referralCode); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	referralPoint := &ReferralPoint{
		App:    appID,
		UserID: userID,
		Points: 10,
	}
	if err := daoManager.ReferralPointDAO.Create(ctx, referralPoint); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	if err := daoManager.ReferralLogDAO.Create(ctx, &ReferralLog{
		UserID:          userID,
		App:             appID,
		ReferralPointID: referralPoint.ID,
		Timestamp:       time.Now().UnixMilli(),
		Type:            consts.ReferralLogTypeGain,
		Reason:          consts.ReferralLogReasonNewUserBilled,
		Points:          10,
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	if err := daoManager.FreeTrialStateDAO.Create(ctx, &FreeTrialState{
		App:       appID,
		UserID:    userID,
		StartDate: now.UnixMilli(),
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	return nil
}

func (s *AlchemistService) JoinReferralProgram(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*alchemist_api.JoinReferralProgramResponse_Data, *gerror.AppError) {
	var err error
	now := time.Now()
	var referralCode *ReferralCode
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		referralCode, err = daoManager.ReferralCodeDAO.GetByUserIDAndApp(ctx, userID, appID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if referralCode == nil {
			code := util.GenerateReferralCode()
			// check wether the generated code exists
			record, err := daoManager.ReferralCodeDAO.GetByCode(ctx, code)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
			if record != nil {
				code = util.GenerateReferralCode()
			}
			record, err = daoManager.ReferralCodeDAO.GetByCode(ctx, code)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
			if record != nil {
				return errors.New("generated referral code exists twice")
			}
			referralCode = &ReferralCode{
				UserID:       userID,
				App:          appID,
				JoinDate:     time.Now(),
				ReferralCode: code,
			}
			return joinReferralProgram(ctx, rpcCtx, now, daoManager, referralCode, appID, userID)
		} else {
			return fmt.Errorf("user [id=%d] has joined the app [%s]", userID, appID)
		}
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
	}
	return &alchemist_api.JoinReferralProgramResponse_Data{
		ReferralCode: referralCode.ReferralCode,
	}, nil
}

func (s *AlchemistService) CheckReferralCode(ctx context.Context, rpcCtx *rpc.Context, code string) (*alchemist_api.CheckReferralCodeResponse_Data, *gerror.AppError) {
	record, err := s.daoManager.ReferralCodeDAO.GetByCode(ctx, code)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record == nil {
		return &alchemist_api.CheckReferralCodeResponse_Data{Valid: false}, nil
	}
	return &alchemist_api.CheckReferralCodeResponse_Data{Valid: true}, nil
}

func (s *AlchemistService) checkReferralCode(ctx context.Context, rpcCtx *rpc.Context, code string) (int64, *gerror.AppError) {
	codeData, err := s.daoManager.ReferralCodeDAO.GetByCode(ctx, code)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if codeData == nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidReferralCode")).WithCode(response.StatusCodeBadRequest)
	}
	return codeData.UserID, nil
}

func (s *AlchemistService) checkUserRegisteredOnOldDeviceRecord(ctx context.Context, rpcCtx *rpc.Context, appID, code string, userID int64) *gerror.AppError {
	record, err := s.daoManager.UserRegisteredOnOldDeviceDAO.GetByUserIDWithAppAndCode(ctx, userID, appID, code)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record != nil {
		err = errors.New("user had registered on old device")
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserRegisteredOnOldDevice")).WithCode(response.StatusCodeUserRegisteredOnOldDevice)
	}
	return nil
}

func upsertReferralNewUser(ctx context.Context, rpcCtx *rpc.Context, daoManager *dao.Manager, now time.Time, appID, code string, userID int64) error {
	record, err := daoManager.ReferralNewUserDAO.GetByUserIDAndAppAndReferralCode(ctx, userID, appID, code)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	if record == nil {
		referralNewUser := &ReferralNewUser{
			App:          appID,
			UserID:       userID,
			BindDate:     now.UnixMilli(),
			ExpiredDate:  now.AddDate(0, 3, 0).UnixMilli(),
			ReferralCode: code,
		}
		if err := daoManager.ReferralNewUserDAO.Create(ctx, referralNewUser); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
	}
	return nil
}

func (s *AlchemistService) BindReferralCode(ctx context.Context, rpcCtx *rpc.Context, appID, code string, userID int64) *gerror.AppError {
	// check the referral code
	referralCodeUserID, appError := s.checkReferralCode(ctx, rpcCtx, code)
	if appError != nil {
		return appError
	}
	// check user_registered_on_old_device record
	if !util.AppConfig(appID).IgnoreDeviceCheck {
		if appError := s.checkUserRegisteredOnOldDeviceRecord(ctx, rpcCtx, appID, code, userID); appError != nil {
			return appError
		}
	}
	// check slark_users record
	now := time.Now()
	slarkUser, err := s.daoManager.SlarkUserDAO.GetByUserID(ctx, userID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if slarkUser != nil {
		if slarkUser.RegisteredAt <= 0 {
			err = fmt.Errorf("user with user id [%d] and app account token [%s], has not registered", slarkUser.UserID, slarkUser.AppAccountToken)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNotRegistered")).WithCode(response.StatusCodeUserHasNotRegistered)
		} else {
			expiration := util.AppConfig(appID).BindReferralCodeExpiration
			if time.UnixMilli(slarkUser.RegisteredAt).AddDate(0, 0, expiration).Before(now) {
				err = fmt.Errorf("user with user id [%d] and app account token [%s], registeration duration exceeds %d days", slarkUser.UserID, slarkUser.AppAccountToken, expiration)
				rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasRegisteredExceed")).WithCode(response.StatusCodeUserHasRegisteredExceed)
			}
		}
	}

	// bind referral code
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// get user record on the app
		if err := upsertReferralNewUser(ctx, rpcCtx, daoManager, now, appID, code, userID); err != nil {
			rpcCtx.Logger.Error("upsertReferralNewUser error", zap.String("method", "BindReferralCode"), zap.NamedError("appError", err))
			return err
		}
		referralPoint, err := daoManager.ReferralPointDAO.GetByUserIDAndApp(ctx, referralCodeUserID, appID)
		if err != nil {
			rpcCtx.Logger.Error("ReferralPointDAO.GetByUserIDAndApp error", zap.String("method", "BindReferralCode"), zap.NamedError("appError", err))
			return err
		}
		if referralPoint == nil {
			rpcCtx.Logger.Error("Referral Code Owner don't have a referral point record", zap.Int64("referralCodeUserID", referralCodeUserID), zap.String("method", "BindReferralCode"))
			return fmt.Errorf("user[id=%d] has no referral point record", userID)
		}
		point := referralPoint.Points
		referralLog := &ReferralLog{
			App:             appID,
			UserID:          referralCodeUserID,
			Timestamp:       now.UnixMilli(),
			Type:            consts.ReferralLogTypeGain,
			ReferralPointID: referralPoint.ID,
		}
		referralTimes, err := daoManager.ReferralLogDAO.CountFirstTimeReferral(ctx, referralCodeUserID)
		if err != nil {
			rpcCtx.Logger.Error("ReferralLogDAO.CountFirstTimeReferral error", zap.String("method", "BindReferralCode"), zap.NamedError("appError", err))
			return err
		}
		if referralTimes == 0 {
			referralLog.Reason = consts.ReferralLogReasonNewUserFirstTime
			referralLog.Points = consts.ReferralPointNewUser
			point += consts.ReferralPointNewUser
		} else {
			referralLog.Reason = consts.ReferralLogReasonNewUser
			referralLog.Points = consts.ReferralPointNonNewUser
			point += consts.ReferralPointNonNewUser
		}
		if err = daoManager.ReferralPointDAO.UpdatePointByID(ctx, referralPoint.ID, point); err != nil {
			rpcCtx.Logger.Error("ReferralPointDAO.UpdatePointByID error", zap.String("method", "BindReferralCode"), zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.ReferralLogDAO.Create(ctx, referralLog); err != nil {
			rpcCtx.Logger.Error("ReferralLogDAO.Create error", zap.String("method", "BindReferralCode"), zap.NamedError("appError", err))
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistService) GetRewardPoints(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*alchemist_api.GetRewardPointsResponse_Data, *gerror.AppError) {
	referralPoint, err := s.daoManager.ReferralPointDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if referralPoint == nil {
		return nil, nil
	}
	list, err := s.daoManager.ReferralLogDAO.GetListByReferralPointID(ctx, referralPoint.ID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var records []*alchemist_api.GetRewardPointsResponse_Record
	var numberOfNewUsers int32
	var numberOfBilledUsers int32
	for _, l := range list {
		records = append(records, &alchemist_api.GetRewardPointsResponse_Record{
			Timestamp: l.Timestamp,
			Type:      l.Type,
			Reason:    consts.ReferralLogReason(rpcCtx.Localizer, l.Reason),
			Points:    l.Points,
		})
		if l.Reason == consts.ReferralLogReasonNewUser {
			numberOfNewUsers++
		} else if l.Reason == consts.ReferralLogReasonNewUserBilled {
			numberOfBilledUsers++
		}
	}
	return &alchemist_api.GetRewardPointsResponse_Data{
		Points:              referralPoint.Points,
		Records:             records,
		NumberOfNewUsers:    numberOfNewUsers,
		NumberOfBilledUsers: numberOfBilledUsers,
	}, nil
}

func (s *AlchemistService) GetNewUserDiscountState(ctx context.Context, rpcCtx *rpc.Context, appID string, billedTimes int32, userID int64) (*alchemist_api.GetNewUserDiscountStateResponse_Data, *gerror.AppError) {
	newUserDiscountState, appError := s.getNewUserDiscountState(ctx, rpcCtx, appID, billedTimes, userID)
	if appError != nil {
		return nil, appError
	}
	var res *alchemist_api.GetNewUserDiscountStateResponse_Data
	if newUserDiscountState != nil {
		res = &alchemist_api.GetNewUserDiscountStateResponse_Data{
			HasNewUserDiscount: newUserDiscountState.HasNewUserDiscount,
			Redeemed:           newUserDiscountState.Redeemed,
			RemainingTimes:     newUserDiscountState.RemainingTimes,
		}
	}
	return res, nil
}

func (s *AlchemistService) getAppAccountToken(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*SlarkUser, *gerror.AppError) {
	accountToken, err := s.daoManager.SlarkUserDAO.GetByUserID(ctx, userID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if accountToken == nil {
		err = fmt.Errorf("user[id=%d] has no app account token", userID)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoAppAccountToken")).WithCode(response.StatusCodeBadRequest)

	}
	return accountToken, nil
}

func (s *AlchemistService) checkNewUserDiscountState(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*NewUserDiscountState, *gerror.AppError) {
	var referralNewUser *ReferralNewUser
	discountState, err := s.daoManager.NewUserDiscountStateDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if discountState == nil {
		referralNewUser, err = s.daoManager.ReferralNewUserDAO.GetByUserIDAndApp(ctx, userID, appID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if referralNewUser == nil {
			err = errors.New("no new-user discount state log and new-user referral log")
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoNewUserDiscountStateAndNoNewUserReferral")).WithCode(response.StatusCodeBadRequest)

		}
		if expiredDate := time.UnixMilli(referralNewUser.ExpiredDate); time.Now().After(expiredDate) {
			err = errors.New("no new-user discount state log and new-user referral expired")
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoNewUserDiscountStateAndExpiredNewUserReferral")).WithCode(response.StatusCodeBadRequest)
		}
	}
	if discountState.RemainingTimes <= 0 {
		err = fmt.Errorf("new-user discount state's remaining times [%d] not greater than zero", discountState.RemainingTimes)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_NewUserDiscountStateRemainingTimeNotGreaterThanZero")).WithCode(response.StatusCodeBadRequest)
	}
	return discountState, nil
}

func (s *AlchemistService) getOfferIDByDiscountState(ctx context.Context, rpcCtx *rpc.Context, appID string, discountState *NewUserDiscountState) (string, *gerror.AppError) {
	var offerID string
	if discountState != nil {
		switch discountState.RemainingTimes {
		case 12:
			offerID = util.AppConfig(appID).DiscountOffer.IDNewUser
		case 10, 11:
			offerID = util.AppConfig(appID).DiscountOffer.ID10M
		case 8, 9:
			offerID = util.AppConfig(appID).DiscountOffer.ID8M
		case 6, 7:
			offerID = util.AppConfig(appID).DiscountOffer.ID6M
		case 4, 5:
			offerID = util.AppConfig(appID).DiscountOffer.ID4M
		case 2, 3:
			offerID = util.AppConfig(appID).DiscountOffer.ID2M
		default:
			err := fmt.Errorf("new-user discount state's remaining times [%d] mismatch", discountState.RemainingTimes)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_NewUserDiscountStateRemainingTimeMismatch")).WithCode(response.StatusCodeBadRequest)
		}
	} else {
		// referralNewUser != nil -- 用户 有优惠，但未使用
		offerID = util.AppConfig(appID).DiscountOffer.IDNewUser
	}
	return offerID, nil
}

func (s *AlchemistService) GetNewUserDiscountSignature(ctx context.Context, rpcCtx *rpc.Context, appID, environment string, userID int64) (*alchemist_api.GetNewUserDiscountSignatureResponse_Data, *gerror.AppError) {
	accountToken, appError := s.getAppAccountToken(ctx, rpcCtx, userID)
	if appError != nil {
		return nil, appError
	}
	// check new-user discount state
	discountState, appError := s.checkNewUserDiscountState(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	// get offerID
	offerID, appError := s.getOfferIDByDiscountState(ctx, rpcCtx, appID, discountState)
	if appError != nil {
		return nil, appError
	}

	// https://developer.apple.com/documentation/storekit/in-app_purchase/original_api_for_in-app_purchase/subscriptions_and_offers/generating_a_signature_for_promotional_offers#3149026
	nonce := util.NewUUIDString()
	timestamp := time.Now().UnixMilli()
	signature, err := util.GeneratePromotionalOfferSignature(appID, util.AppConfig(appID).AppID, util.AppConfig(appID).PromoOfferKeyID, util.AppConfig(appID).AppID, offerID, accountToken.AppAccountToken, nonce, strconv.FormatInt(timestamp, 10))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	if err := s.daoManager.PromoOfferRecordsDAO.Create(ctx, &PromoOfferRecord{
		App:         appID,
		UserID:      userID,
		OfferID:     offerID,
		SignDate:    timestamp,
		Environment: consts.EnvironmentNum(strings.TrimSpace(environment)),
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	return &alchemist_api.GetNewUserDiscountSignatureResponse_Data{
		OfferID:            offerID,
		CanConvertToPoints: discountState.RemainingTimes%2 == 1,
		SignatureInfo: &alchemist_api.GetNewUserDiscountSignatureResponse_SignatureInfo{
			KeyID:     util.AppConfig(appID).PromoOfferKeyID,
			Nonce:     nonce,
			Timestamp: timestamp,
			Signature: signature,
		},
	}, nil
}

func (s *AlchemistService) UseNewUserDiscountOffer(ctx context.Context, rpcCtx *rpc.Context, appID, offerID, environment string, userID int64) *gerror.AppError {
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		var newUserDiscountState *NewUserDiscountState
		newUserDiscountState, err := daoManager.NewUserDiscountStateDAO.GetByUserID(ctx, userID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if newUserDiscountState == nil {
			// 如果是新用户第一次使用： 则新建记录
			var referralCode *ReferralCode
			referralCode, err := daoManager.ReferralCodeDAO.GetByUserIDAndApp(ctx, userID, appID)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
			if referralCode == nil {
				err = fmt.Errorf("user[id=%d] has no referral code: %w", userID, err)
				rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
				return err
			}
			newUserDiscountState = &NewUserDiscountState{
				App:            appID,
				UserID:         userID,
				ReferralCode:   referralCode.ReferralCode,
				StartDate:      time.Now().UnixMilli(),
				BilledTimes:    0, // 初始值是0和12 接收到notification，才更新， 从0到1，才发推广付费用户的奖励。
				RemainingTimes: 12,
			}
			if err = daoManager.NewUserDiscountStateDAO.Create(ctx, newUserDiscountState); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
		} else {
			// newUserDiscountState.BilledTimes += 1
			// newUserDiscountState.RemainingTimes -= 1
			if err = daoManager.NewUserDiscountStateDAO.UpdateByID(ctx, newUserDiscountState.ID, newUserDiscountState); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
		}
		// update promo_offer_records
		promoOfferRecord, err := daoManager.PromoOfferRecordsDAO.GetByUserIDAndOfferIDAppEnv(ctx, userID, offerID, appID, consts.EnvironmentNum(strings.TrimSpace(environment)))
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if promoOfferRecord == nil {
			err = fmt.Errorf("user [id=%d] has no promotional offer record of app[%s] with offerID[%s] in environment[%s]", userID, appID, offerID, environment)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return err
		}
		promoOfferRecord.UseDate = time.Now().UnixMilli()
		if err = daoManager.PromoOfferRecordsDAO.UpdateByID(ctx, promoOfferRecord.ID, promoOfferRecord); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
	}
	return nil
}

func (s *AlchemistService) GetTrialState(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*alchemist_api.GetTrialStateResponse_Data, *gerror.AppError) {
	freeTrialState, appError := s.getFreeTrialState(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	var res *alchemist_api.GetTrialStateResponse_Data
	if freeTrialState != nil {
		res = &alchemist_api.GetTrialStateResponse_Data{
			InFreeTrial:      freeTrialState.InFreeTrial,
			ExpirationDate:   freeTrialState.ExpirationDate,
			StartDate:        freeTrialState.StartDate,
			DaysOfTrial:      freeTrialState.DaysOfTrial,
			TotalDaysOfTrial: freeTrialState.TotalDaysOfTrial,
		}
	}
	return res, nil
}

func (s *AlchemistService) GetRewardList(ctx context.Context, rpcCtx *rpc.Context, appID string) (*alchemist_api.GetRewardListResponse_Data, *gerror.AppError) {
	var list []*alchemist_api.GetRewardListResponse_Reward
	for _, reward := range util.AppConfig(appID).RewardList {
		list = append(list, &alchemist_api.GetRewardListResponse_Reward{
			Id:             reward.ID,
			Name:           reward.Name,
			Description:    reward.Description,
			OfferID:        reward.OfferID,
			Cost:           reward.Cost,
			Duration:       reward.Duration,
			DurationInDays: reward.DurationInDays,
		})
	}
	return &alchemist_api.GetRewardListResponse_Data{RewardList: list}, nil
}

func (s *AlchemistService) getReferralPoint(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*ReferralPoint, *gerror.AppError) {
	referralPoint, err := s.daoManager.ReferralPointDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if referralPoint == nil {
		err = fmt.Errorf("user [id=%d] has no referral point record", userID)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoReferralPointRecord")).WithCode(response.StatusCodeBadRequest)

	}
	return referralPoint, nil
}

func (s *AlchemistService) checkReward(ctx context.Context, rpcCtx *rpc.Context, appID, rewardID string, referralPoint *ReferralPoint) (*util.PromoReward, *gerror.AppError) {
	reward, err := util.Reward(appID, rewardID)
	if err != nil {
		if err != util.ErrorInvalidRewardId {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidRewardId")).WithCode(response.StatusCodeBadRequest)
		}
	}
	if referralPoint.Points < reward.Cost {
		err = fmt.Errorf("referral points [%d] not enough to cost [%d]", referralPoint.Points, reward.Cost)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ReferalPointsNotEnoughToCost")).WithCode(response.StatusCodeBadRequest)
	}
	return reward, nil
}

func (s *AlchemistService) RedeemReward(ctx context.Context, rpcCtx *rpc.Context, appID, rewardID string, userID int64) (*alchemist_api.RedeemRewardResponse_Data, *gerror.AppError) {
	// get referral point
	referralPoint, appError := s.getReferralPoint(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	reward, appError := s.checkReward(ctx, rpcCtx, appID, rewardID, referralPoint)
	if appError != nil {
		return nil, appError
	}
	var freeTrialState *FreeTrialState
	now := time.Now()
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		err := daoManager.ReferralPointDAO.UpdatePointByID(ctx, referralPoint.ID, referralPoint.Points-reward.Cost)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.ReferralLogDAO.Create(ctx, &ReferralLog{
			ReferralPointID: referralPoint.ID,
			App:             appID,
			UserID:          userID,
			Timestamp:       time.Now().UnixMilli(),
			Type:            consts.ReferralLogTypeConsume,
			Reason:          consts.ReferralLogReasonRedeemReward,
			Points:          reward.Cost,
		}); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		freeTrialState, err = daoManager.FreeTrialStateDAO.GetByUserIDAndApp(ctx, userID, appID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if freeTrialState == nil {
			err = fmt.Errorf("user[id=%d] has no free trial state log of app[%s]", userID, appID)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return err
		}
		if expiratoinDate := time.UnixMilli(freeTrialState.ExpirationDate); expiratoinDate.After(now) {
			// 有未结束的试用期
			freeTrialState.ExpirationDate = expiratoinDate.AddDate(0, 0, int(reward.DurationInDays)).UnixMilli()
			freeTrialState.DaysOfTrial += reward.DurationInDays
			freeTrialState.TotalDaysOfTrial += reward.DurationInDays
		} else {
			freeTrialState.StartDate = now.UnixMilli()
			freeTrialState.ExpirationDate = now.AddDate(0, 0, int(reward.DurationInDays)).UnixMilli()
			freeTrialState.DaysOfTrial = reward.DurationInDays
			freeTrialState.TotalDaysOfTrial += reward.DurationInDays
		}
		if err = daoManager.FreeTrialStateDAO.UpdateByID(ctx, freeTrialState.ID, freeTrialState); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		return nil
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &alchemist_api.RedeemRewardResponse_Data{
		RemainingPoints: referralPoint.Points - reward.Cost,
		ExpirationDate:  freeTrialState.ExpirationDate,
	}, nil
}

func (s *AlchemistService) RedeemRewardOffer(ctx context.Context, rpcCtx *rpc.Context, appID, rewardID, environment string, userID int64) (*alchemist_api.RedeemRewardOfferResponse_Data, *gerror.AppError) {
	accountToken, appError := s.getAppAccountToken(ctx, rpcCtx, userID)
	if appError != nil {
		return nil, appError
	}
	// get referral point
	referralPoint, appError := s.getReferralPoint(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	reward, appError := s.checkReward(ctx, rpcCtx, appID, rewardID, referralPoint)
	if appError != nil {
		return nil, appError
	}

	nonce := util.NewUUIDString()
	timestamp := time.Now().UnixMilli()
	signature, err := util.GeneratePromotionalOfferSignature(appID, util.AppConfig(appID).AppID, util.AppConfig(appID).PromoOfferKeyID, util.AppConfig(appID).AppID, reward.OfferID, accountToken.AppAccountToken, nonce, strconv.FormatInt(timestamp, 10))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	if err := s.daoManager.PromoOfferRecordsDAO.Create(ctx, &PromoOfferRecord{
		App:         appID,
		UserID:      userID,
		OfferID:     reward.OfferID,
		SignDate:    time.Now().UnixMilli(),
		Environment: consts.EnvironmentNum(strings.TrimSpace(environment)),
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	return &alchemist_api.RedeemRewardOfferResponse_Data{
		OfferID: reward.OfferID,
		SignatureInfo: &alchemist_api.RedeemRewardOfferResponse_SignatureInfo{
			KeyID:     util.AppConfig(appID).PromoOfferKeyID,
			Nonce:     nonce,
			Timestamp: timestamp,
			Signature: signature,
		},
	}, nil
}

func (s *AlchemistService) FinishRewardOffer(ctx context.Context, rpcCtx *rpc.Context, appID, rewardID, environment string, userID int64) (*alchemist_api.FinishRewardOfferResponse_Data, *gerror.AppError) {
	// get referral point
	referralPoint, appError := s.getReferralPoint(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	reward, appError := s.checkReward(ctx, rpcCtx, appID, rewardID, referralPoint)
	if appError != nil {
		return nil, appError
	}
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		var err error
		var freeTrialState *FreeTrialState
		freeTrialState, err = daoManager.FreeTrialStateDAO.GetByUserIDAndApp(ctx, userID, appID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if freeTrialState == nil {
			err = fmt.Errorf("user[id=%d] has no free trial state log of app[%s]", userID, appID)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.FreeTrialStateDAO.UpdateColumnByID(
			ctx,
			freeTrialState.ID,
			"total_days_of_trial",
			freeTrialState.TotalDaysOfTrial+reward.DurationInDays,
		); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.ReferralPointDAO.UpdatePointByID(ctx, referralPoint.ID, referralPoint.Points-reward.Cost); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.ReferralLogDAO.Create(ctx, &ReferralLog{
			ReferralPointID: referralPoint.ID,
			App:             appID,
			UserID:          userID,
			Timestamp:       time.Now().UnixMilli(),
			Type:            consts.ReferralLogTypeConsume,
			Reason:          consts.ReferralLogReasonRedeemReward,
			Points:          reward.Cost,
		}); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		var promoOfferRecord *PromoOfferRecord
		promoOfferRecord, err = daoManager.PromoOfferRecordsDAO.GetByUserIDAndOfferIDAppEnv(
			ctx,
			userID,
			reward.OfferID,
			appID,
			consts.EnvironmentNum(environment),
		)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if promoOfferRecord == nil {
			err = fmt.Errorf("user[id=%d] has no promotional offer record of app[%s] with offer id [%s] in %s environment", userID, appID, reward.OfferID, environment)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.PromoOfferRecordsDAO.UpdateByID(ctx, promoOfferRecord.ID, &PromoOfferRecord{UseDate: time.Now().UnixMilli()}); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		return nil
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &alchemist_api.FinishRewardOfferResponse_Data{RemainingPoints: referralPoint.Points - reward.Cost}, nil
}

func (s *AlchemistService) ConvertUnusedNewUserDiscount(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) *gerror.AppError {
	discountState, err := s.daoManager.NewUserDiscountStateDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if discountState == nil {
		err = fmt.Errorf("user[id=%d] has no new-user discount state record of app[%s]", userID, appID)
		rpcCtx.Logger.Error("bad requeset", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoNewUserDiscountState")).WithCode(response.StatusCodeBadRequest)

	}
	if discountState.RemainingTimes != 1 {
		err = fmt.Errorf("user[id=%d], new-user discount state's remaining times[%d] of app[%s] not equal to 1",
			userID, discountState.RemainingTimes, appID)
		rpcCtx.Logger.Error("bad requeset", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_NewUserDiscountStateRemainingTimeNotEqualToOne")).WithCode(response.StatusCodeBadRequest)
	}
	// get referral point
	referralPoint, appError := s.getReferralPoint(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return appError
	}
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		if err = daoManager.NewUserDiscountStateDAO.UpdateColumnByID(ctx, discountState.ID, "remaining_times", 0); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.ReferralPointDAO.UpdatePointByID(ctx, referralPoint.ID, referralPoint.Points+4); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if err = daoManager.ReferralLogDAO.Create(ctx, &ReferralLog{
			ReferralPointID: referralPoint.ID,
			UserID:          userID,
			Timestamp:       time.Now().UnixMilli(),
			Type:            consts.ReferralLogTypeGain,
			Reason:          consts.ReferralLogReasonConvertPoints,
			Points:          4,
			App:             appID,
		}); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistService) CheckReferralCodeAndDevice(ctx context.Context, rpcCtx *rpc.Context, appID, deviceToken, referralCode string, userID int64) (*alchemist_api.CheckReferralCodeAndDeviceResponse_Data, *gerror.AppError) {
	// check code
	record, err := s.daoManager.ReferralCodeDAO.GetByCode(ctx, referralCode)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record == nil {
		return &alchemist_api.CheckReferralCodeAndDeviceResponse_Data{
			CodeValid:   false,
			DeviceValid: false,
		}, nil
	}
	// check user_registered_on_old_device record
	if !util.AppConfig(appID).IgnoreDeviceCheck {
		record, err := s.daoManager.UserRegisteredOnOldDeviceDAO.GetByUserIDWithAppAndCode(ctx, userID, appID, referralCode)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if record != nil {
			return &alchemist_api.CheckReferralCodeAndDeviceResponse_Data{
				CodeValid:   true,
				DeviceValid: false,
			}, nil
		}
	}

	// device check
	var result devicecheck.QueryTwoBitsResult
	var dcEnv devicecheck.Environment
	switch s.env {
	case gutil.AppEnvDEV, gutil.AppEnvPPE:
		dcEnv = devicecheck.Development
	case gutil.AppEnvPROD:
		dcEnv = devicecheck.Production
	}
	cred := devicecheck.NewCredentialString(util.AppConfig(appID).DeviceCheck.PrivKeyPem)
	cfg := devicecheck.NewConfig(util.AppConfig(appID).DeviceCheck.IssuerID, util.AppConfig(appID).DeviceCheck.KeyID, dcEnv)
	if err := devicecheck.New(cred, cfg).QueryTwoBits(ctx, deviceToken, &result); err != nil && !errors.Is(err, devicecheck.ErrBitStateNotFound) {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
	}

	if result.Bit1 {
		if err := s.daoManager.UserRegisteredOnOldDeviceDAO.Create(ctx, &UserRegisteredOnOldDevice{
			App:          appID,
			UserID:       userID,
			IP:           rpcCtx.IP,
			ReferralCode: referralCode,
		}); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		return &alchemist_api.CheckReferralCodeAndDeviceResponse_Data{
			CodeValid:   true,
			DeviceValid: false,
		}, nil
	}

	// rpcCtx.Logger.Info("valid referral code and device", zap.String("referral code", referralCode), zap.String("device token", deviceToken))
	return &alchemist_api.CheckReferralCodeAndDeviceResponse_Data{
		CodeValid:   true,
		DeviceValid: true,
	}, nil
}

func (s *AlchemistService) MarkNewRegistration(ctx context.Context, rpcCtx *rpc.Context, appID, deviceToken0, deviceToken1 string, userID int64) *gerror.AppError {
	var result devicecheck.QueryTwoBitsResult
	var dcEnv devicecheck.Environment
	switch s.env {
	case gutil.AppEnvDEV, gutil.AppEnvPPE:
		dcEnv = devicecheck.Development
	case gutil.AppEnvPROD:
		dcEnv = devicecheck.Production
	}
	cred := devicecheck.NewCredentialString(util.AppConfig(appID).DeviceCheck.PrivKeyPem)
	cfg := devicecheck.NewConfig(util.AppConfig(appID).DeviceCheck.IssuerID, util.AppConfig(appID).DeviceCheck.KeyID, dcEnv)
	dck := devicecheck.New(cred, cfg)
	if err := dck.QueryTwoBits(ctx, deviceToken0, &result); err != nil && !errors.Is(err, devicecheck.ErrBitStateNotFound) {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
	}
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// upsert slark_users
		accountToken, err := daoManager.SlarkUserDAO.GetByUserID(ctx, userID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		if accountToken == nil {
			accountToken = &SlarkUser{
				AppAccountToken: util.NewUUIDString(),
				RegisteredAt:    time.Now().UnixMilli(),
				UserID:          userID,
			}
			if err := daoManager.SlarkUserDAO.Create(ctx, accountToken); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
		}
		if accountToken.RegisteredAt <= 0 {
			if err := daoManager.SlarkUserDAO.UpdateColumnByID(ctx, accountToken.ID, "registered_at", time.Now().UnixMilli()); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
		}
		// handle device check result
		if result.Bit1 {
			err := errors.New("devicecheck result: no need to update")
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return err
		} else {
			if err := dck.UpdateTwoBits(ctx, deviceToken1, result.Bit0, true); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistService) Unregister(ctx context.Context, rpcCtx *rpc.Context, userID int64) *gerror.AppError {
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		var err error
		// delete account_tokens
		if err = daoManager.SlarkUserDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete subscription_state_prod
		if err = daoManager.SubscriptionStateProdDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete subscription_state_test
		if err = daoManager.SubscriptionStateTestDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete transactions_prod
		if err = daoManager.TransactionsProdDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete transactions_test
		if err = daoManager.TransactionsTestDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete referral_code
		if err = daoManager.ReferralCodeDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete referral_new_user
		if err = daoManager.ReferralNewUserDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete referral_point
		if err = daoManager.ReferralPointDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete referral_log
		if err = daoManager.ReferralLogDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete promo_offer_records
		if err = daoManager.PromoOfferRecordsDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete new_user_discount_state
		if err = daoManager.NewUserDiscountStateDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete free_trial_state
		if err = daoManager.FreeTrialStateDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// delete user_registered_on_old_device
		if err = daoManager.UserRegisteredOnOldDeviceDAO.DeleteByUserID(ctx, userID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistService) getNewUserStatusExpirationDate(ctx context.Context, rpcCtx *rpc.Context, appID, myReferralCode string, userID int64) (int64, *gerror.AppError) {
	var newUserStatusExpirationdate int64 = 0
	if myReferralCode == "" {
		// check user_registered_on_old_device record
		record, err := s.daoManager.UserRegisteredOnOldDeviceDAO.GetByUserIDWithApp(ctx, userID, appID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return newUserStatusExpirationdate, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if record == nil {
			// check slark_users
			slarkUser, err := s.daoManager.SlarkUserDAO.GetByUserID(ctx, userID)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return newUserStatusExpirationdate, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if slarkUser == nil {
				err = fmt.Errorf("slark user [%d] data missing", userID)
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return newUserStatusExpirationdate, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoAppAccountToken")).WithCode(response.StatusCodeBadRequest)
			}
			if slarkUser.RegisteredAt <= 0 {
				err = fmt.Errorf("user with user id [%d] and app account token [%s], has not registered", slarkUser.UserID, slarkUser.AppAccountToken)
				rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
				return newUserStatusExpirationdate, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNotRegistered")).WithCode(response.StatusCodeUserHasNotRegistered)
			} else {
				expiration := util.AppConfig(appID).BindReferralCodeExpiration
				newUserStatusExpirationdate = time.UnixMilli(slarkUser.RegisteredAt).AddDate(0, 0, expiration).UnixMilli()
			}
		}
	}
	return newUserStatusExpirationdate, nil
}

func (s *AlchemistService) getReferralCodeState(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*alchemist_api.GetStateResponse_ReferralCodeState, *gerror.AppError) {
	// my referral code
	var myReferralCode string
	referralCode, err := s.daoManager.ReferralCodeDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if referralCode != nil {
		myReferralCode = referralCode.ReferralCode
	}
	// new user status expiration date
	newUserStatusExpirationdate, appError := s.getNewUserStatusExpirationDate(ctx, rpcCtx, appID, myReferralCode, userID)
	if appError != nil {
		return nil, appError
	}
	// share url
	shareURL := ""
	switch s.env {
	case gutil.AppEnvDEV, gutil.AppEnvPPE:
		shareURL = "http://rp.test.n1xt.net/s/" + myReferralCode
	case gutil.AppEnvPROD:
		shareURL = "http://rp.n1xt.net/s/" + myReferralCode
	}
	// used referral code
	usedReferralCode := ""
	referralNewUser, err := s.daoManager.ReferralNewUserDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if referralNewUser != nil {
		usedReferralCode = referralNewUser.ReferralCode
	}

	var myReferralCodeVal = structpb.NewNullValue()
	if myReferralCode != "" {
		myReferralCodeVal = structpb.NewStringValue(myReferralCode)
	}
	var newUserStatusExpirationdateVal = structpb.NewNullValue()
	if newUserStatusExpirationdate > 0 {
		newUserStatusExpirationdateVal = structpb.NewNumberValue(float64(newUserStatusExpirationdate))
	}
	var usedReferralCodeVal = structpb.NewNullValue()
	if usedReferralCode != "" {
		usedReferralCodeVal = structpb.NewStringValue(usedReferralCode)
	}

	structpb.NewValue(structpb.NullValue_NULL_VALUE)
	return &alchemist_api.GetStateResponse_ReferralCodeState{
		MyReferralCode:              myReferralCodeVal,
		NewUserStatusExpirationDate: newUserStatusExpirationdateVal,
		ShareURL:                    shareURL,
		UsedReferralCode:            usedReferralCodeVal,
	}, nil
}

type freeTrialState struct {
	InFreeTrial      bool
	ExpirationDate   int64
	StartDate        int64
	DaysOfTrial      int32
	TotalDaysOfTrial int32
}

func (s *AlchemistService) getFreeTrialState(ctx context.Context, rpcCtx *rpc.Context, appID string, userID int64) (*freeTrialState, *gerror.AppError) {
	trialState, err := s.daoManager.FreeTrialStateDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if trialState == nil {
		rpcCtx.Logger.Warn(fmt.Sprintf("user[id=%d] has no free trial state record", userID))
		return nil, nil
	}
	return &freeTrialState{
		InFreeTrial:      time.UnixMilli(trialState.ExpirationDate).After(time.Now()),
		ExpirationDate:   trialState.ExpirationDate,
		StartDate:        trialState.StartDate,
		DaysOfTrial:      trialState.DaysOfTrial,
		TotalDaysOfTrial: trialState.TotalDaysOfTrial,
	}, nil
}

type newUserDiscountState struct {
	HasNewUserDiscount bool
	Redeemed           bool
	RemainingTimes     int32
}

func (s *AlchemistService) getNewUserDiscountState(ctx context.Context, rpcCtx *rpc.Context, appID string, billedTimes int32, userID int64) (*newUserDiscountState, *gerror.AppError) {
	state := &newUserDiscountState{
		HasNewUserDiscount: true,
		Redeemed:           true,
		RemainingTimes:     12,
	}
	// get new_user_discount_state record
	discountState, err := s.daoManager.NewUserDiscountStateDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// has no new_user_discount_state record
	if discountState == nil {
		referralNewUser, err := s.daoManager.ReferralNewUserDAO.GetByUserIDAndApp(ctx, userID, appID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if referralNewUser == nil {
			rpcCtx.Logger.Warn(fmt.Sprintf("user[id=%d] has no referral new-user record of app[%s]", userID, appID))
			return nil, nil
		}
		// but has referral_new_user record
		if expiredData := time.UnixMilli(referralNewUser.ExpiredDate); time.Now().After(expiredData) {
			// if referral_new_user expires, do warning log
			rpcCtx.Logger.Warn(fmt.Sprintf("user[id=%d] referral new-user record on app[%s] has expired", userID, appID))
			state.HasNewUserDiscount = false
			state.Redeemed = false
			state.RemainingTimes = 0
		} else {
			state.Redeemed = false
		}
	} else {
		// if has new_user_discount_state record
		state.RemainingTimes = discountState.RemainingTimes
		// check remaining times
		if discountState.RemainingTimes <= 0 {
			state.HasNewUserDiscount = false
		}
		// warning billed times
		if discountState.BilledTimes != billedTimes {
			rpcCtx.Logger.Warn(fmt.Sprintf("client billedTimes[%d] not equal to the billedTimes[%d] in db", billedTimes, discountState.BilledTimes))
		}
	}
	return state, nil
}

func (s *AlchemistService) GetState(ctx context.Context, rpcCtx *rpc.Context, appID string, billedTimes int32, userID int64) (*alchemist_api.GetStateResponse_Data, *gerror.AppError) {
	var data alchemist_api.GetStateResponse_Data
	// referral code state
	referralCodeState, appError := s.getReferralCodeState(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	data.ReferralCodeState = referralCodeState
	// reward points
	referralPoint, err := s.daoManager.ReferralPointDAO.GetByUserIDAndApp(ctx, userID, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if referralPoint != nil {
		data.RewardPoints = referralPoint.Points
	}
	// free trial state
	freeTrialState, appError := s.getFreeTrialState(ctx, rpcCtx, appID, userID)
	if appError != nil {
		return nil, appError
	}
	if freeTrialState != nil {
		data.TrialState = &alchemist_api.GetStateResponse_TrialState{
			InFreeTrial:      freeTrialState.InFreeTrial,
			ExpirationDate:   freeTrialState.ExpirationDate,
			StartDate:        freeTrialState.StartDate,
			DaysOfTrial:      freeTrialState.DaysOfTrial,
			TotalDaysOfTrial: freeTrialState.TotalDaysOfTrial,
		}
	}
	// new user discount state
	newUserDiscountState, appError := s.getNewUserDiscountState(ctx, rpcCtx, appID, billedTimes, userID)
	if appError != nil {
		return nil, appError
	}
	if newUserDiscountState != nil {
		data.NewUserDiscountState = &alchemist_api.GetStateResponse_NewUserDiscountState{
			HasNewUserDiscount: newUserDiscountState.HasNewUserDiscount,
			Redeemed:           newUserDiscountState.Redeemed,
			RemainingTimes:     newUserDiscountState.RemainingTimes,
		}
	}
	return &data, nil
}
