package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/alchemist/pkg/consts"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func generateTransactionsProd(transaction *util.TransactionInfo, renewalInfo *util.RenewalInfo) *TransactionsProd {
	one := TransactionsProd{
		OriginalTransactionID:       transaction.OriginalTransactionId,
		TransactionID:               transaction.TransactionId,
		WebOrderLineItemID:          transaction.WebOrderLineItemId,
		BundleID:                    transaction.BundleId,
		AppAccountToken:             transaction.AppAccountToken,
		ProductID:                   transaction.ProductId,
		Type:                        transaction.Type,
		SubscriptionGroupIdentifier: transaction.SubscriptionGroupIdentifier,
		Quantity:                    transaction.Quantity,
		Price:                       transaction.Price,
		Currency:                    transaction.Currency,
		Storefront:                  transaction.Storefront,
		StorefrontID:                transaction.StorefrontId,
		OfferIdentifier:             transaction.OfferIdentifier,
		OfferType:                   transaction.OfferType,
		OfferDiscountType:           transaction.OfferDiscountType,
		OriginalPurchaseDate:        transaction.OriginalPurchaseDate,
		PurchaseDate:                transaction.PurchaseDate,
		RecentSubscriptionStartDate: renewalInfo.RecentSubscriptionStartDate,
		IsInBillingRetryPeriod:      renewalInfo.IsInBillingRetryPeriod,
		GracePeriodExpiresDate:      renewalInfo.GracePeriodExpiresDate,
		AutoRenewStatus:             renewalInfo.AutoRenewStatus,
		AutoRenewProductID:          renewalInfo.AutoRenewProductId,
		ExpirationIntent:            renewalInfo.ExpirationIntent,
		ExpiresDate:                 transaction.ExpiresDate,
		IsUpgraded:                  transaction.IsUpgraded,
		RenewalDate:                 renewalInfo.RenewalDate,
		InAppOwnershipType:          transaction.InAppOwnershipType,
		PriceIncreaseStatus:         renewalInfo.PriceIncreaseStatus,
		RevocationDate:              transaction.RevocationDate,
		RevocationReason:            transaction.RevocationReason,
		TransactionReason:           transaction.TransactionReason,
		SignedDate:                  transaction.SignedDate,
	}
	return &one
}

func generateTransactionsTest(transaction *util.TransactionInfo, renewalInfo *util.RenewalInfo) *TransactionsTest {
	one := TransactionsTest{
		OriginalTransactionID:       transaction.OriginalTransactionId,
		TransactionID:               transaction.TransactionId,
		WebOrderLineItemID:          transaction.WebOrderLineItemId,
		BundleID:                    transaction.BundleId,
		AppAccountToken:             transaction.AppAccountToken,
		ProductID:                   transaction.ProductId,
		Type:                        transaction.Type,
		SubscriptionGroupIdentifier: transaction.SubscriptionGroupIdentifier,
		Quantity:                    transaction.Quantity,
		Price:                       transaction.Price,
		Currency:                    transaction.Currency,
		Storefront:                  transaction.Storefront,
		StorefrontID:                transaction.StorefrontId,
		OfferIdentifier:             transaction.OfferIdentifier,
		OfferType:                   transaction.OfferType,
		OfferDiscountType:           transaction.OfferDiscountType,
		OriginalPurchaseDate:        transaction.OriginalPurchaseDate,
		PurchaseDate:                transaction.PurchaseDate,
		RecentSubscriptionStartDate: renewalInfo.RecentSubscriptionStartDate,
		IsInBillingRetryPeriod:      renewalInfo.IsInBillingRetryPeriod,
		GracePeriodExpiresDate:      renewalInfo.GracePeriodExpiresDate,
		AutoRenewStatus:             renewalInfo.AutoRenewStatus,
		AutoRenewProductID:          renewalInfo.AutoRenewProductId,
		ExpirationIntent:            renewalInfo.ExpirationIntent,
		ExpiresDate:                 transaction.ExpiresDate,
		IsUpgraded:                  transaction.IsUpgraded,
		RenewalDate:                 renewalInfo.RenewalDate,
		InAppOwnershipType:          transaction.InAppOwnershipType,
		PriceIncreaseStatus:         renewalInfo.PriceIncreaseStatus,
		RevocationDate:              transaction.RevocationDate,
		RevocationReason:            transaction.RevocationReason,
		TransactionReason:           transaction.TransactionReason,
		SignedDate:                  transaction.SignedDate,
	}
	return &one
}

func insertTransactionsProd(ctx context.Context, daoManager *dao.Manager, accountToken *SlarkUser, transactionInfo *util.TransactionInfo, renewalInfo *util.RenewalInfo) error {
	transaction := generateTransactionsProd(transactionInfo, renewalInfo)
	transaction.UserID = accountToken.UserID
	return daoManager.TransactionsProdDAO.Create(ctx, transaction)
}

func insertTransactionsTest(ctx context.Context, daoManager *dao.Manager, accountToken *SlarkUser, transactionInfo *util.TransactionInfo, renewalInfo *util.RenewalInfo) error {
	transaction := generateTransactionsTest(transactionInfo, renewalInfo)
	transaction.UserID = accountToken.UserID
	return daoManager.TransactionsTestDAO.Create(ctx, transaction)
}

func generateSubscriptionStateProd(transaction *util.TransactionInfo, renewalInfo *util.RenewalInfo) *SubscriptionStateProd {
	one := SubscriptionStateProd{
		FirstTimeSubscribedAt: transaction.PurchaseDate,
		SubscribedAt:          transaction.OriginalPurchaseDate,
		CurrentBillDate:       transaction.PurchaseDate,
		CurrentBillPrice:      decimal.New(int64(transaction.Price), -3).String(),
		CurrentOfferType:      consts.OfferType(transaction.OfferType),
		CurrentOfferID:        transaction.OfferIdentifier,
		ExpirationDate:        transaction.ExpiresDate,
		AutoRenew:             renewalInfo.AutoRenewStatus == 1,
		NextOfferType:         consts.OfferType(renewalInfo.OfferType),
		NextOfferID:           renewalInfo.OfferIdentifier,
		CurrencyCode:          transaction.Currency,
	}
	one.NextBillPrice = one.CurrentBillPrice
	return &one
}

func generateSubscriptionStateTest(transaction *util.TransactionInfo, renewalInfo *util.RenewalInfo) *SubscriptionStateTest {
	one := SubscriptionStateTest{
		FirstTimeSubscribedAt: transaction.PurchaseDate,
		SubscribedAt:          transaction.OriginalPurchaseDate,
		CurrentBillDate:       transaction.PurchaseDate,
		CurrentBillPrice:      decimal.New(int64(transaction.Price), -3).String(),
		CurrentOfferType:      consts.OfferType(transaction.OfferType),
		CurrentOfferID:        transaction.OfferIdentifier,
		ExpirationDate:        transaction.ExpiresDate,
		AutoRenew:             renewalInfo.AutoRenewStatus == 1,
		NextOfferType:         consts.OfferType(renewalInfo.OfferType),
		NextOfferID:           renewalInfo.OfferIdentifier,
		CurrencyCode:          transaction.Currency,
	}
	one.NextBillPrice = one.CurrentBillPrice
	return &one
}

func insertSubscriptionStateProd(ctx context.Context, daoManager *dao.Manager, accountToken *SlarkUser, notificationType string, transactionInfo *util.TransactionInfo, renewalInfo *util.RenewalInfo) error {
	subscriptionState := generateSubscriptionStateProd(transactionInfo, renewalInfo)
	subscriptionState.UserID = accountToken.UserID
	if notificationType == consts.SUBSCRIBED || notificationType == consts.DID_RENEW {
		subscriptionState.Subscribed = true
	} else if notificationType == consts.EXPIRED {
		subscriptionState.Subscribed = false
	}
	return daoManager.SubscriptionStateProdDAO.Create(ctx, subscriptionState)
}

func insertSubscriptionStateTest(ctx context.Context, daoManager *dao.Manager, accountToken *SlarkUser, notificationType string, transactionInfo *util.TransactionInfo, renewalInfo *util.RenewalInfo) error {
	subscriptionState := generateSubscriptionStateTest(transactionInfo, renewalInfo)
	subscriptionState.UserID = accountToken.UserID
	if notificationType == consts.SUBSCRIBED || notificationType == consts.DID_RENEW {
		subscriptionState.Subscribed = true
	} else if notificationType == consts.EXPIRED {
		subscriptionState.Subscribed = false
	}
	return daoManager.SubscriptionStateTestDAO.Create(ctx, subscriptionState)
}

func parseAppStoreServerNotification(ctx context.Context, daoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction) (*util.AppStoreServerNotification, error) {
	appStoreServerRequest := rawTransaction.Data
	// handle notification data
	var request util.AppStoreServerRequest
	var err error
	// bind byte to header structure
	if err = json.Unmarshal([]byte(appStoreServerRequest), &request); err != nil {
		err = fmt.Errorf("unmarshal AppStoreServerRequest failed: %w", err)
		logger.Warn("appStoreServerRequest unmarshal failed", zap.NamedError("appError", err))
		if err := daoManager.RawTransactionsDAO.UpdateHandleErrorByID(ctx, rawTransaction.ID, err.Error()); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
		}
		return nil, err
	}
	// Apple Root CA - G3 Root certificate
	// for details: https://www.apple.com/certificateauthority/
	// you need download it and covert it to a valid pem file in order to verify X5c certificates
	// `openssl x509 -in AppleRootCA-G3.cer -out cert.pem`
	rootCert := `-----BEGIN CERTIFICATE-----
MIICQzCCAcmgAwIBAgIILcX8iNLFS5UwCgYIKoZIzj0EAwMwZzEbMBkGA1UEAwwS
QXBwbGUgUm9vdCBDQSAtIEczMSYwJAYDVQQLDB1BcHBsZSBDZXJ0aWZpY2F0aW9u
IEF1dGhvcml0eTETMBEGA1UECgwKQXBwbGUgSW5jLjELMAkGA1UEBhMCVVMwHhcN
MTQwNDMwMTgxOTA2WhcNMzkwNDMwMTgxOTA2WjBnMRswGQYDVQQDDBJBcHBsZSBS
b290IENBIC0gRzMxJjAkBgNVBAsMHUFwcGxlIENlcnRpZmljYXRpb24gQXV0aG9y
aXR5MRMwEQYDVQQKDApBcHBsZSBJbmMuMQswCQYDVQQGEwJVUzB2MBAGByqGSM49
AgEGBSuBBAAiA2IABJjpLz1AcqTtkyJygRMc3RCV8cWjTnHcFBbZDuWmBSp3ZHtf
TjjTuxxEtX/1H7YyYl3J6YRbTzBPEVoA/VhYDKX1DyxNB0cTddqXl5dvMVztK517
IDvYuVTZXpmkOlEKMaNCMEAwHQYDVR0OBBYEFLuw3qFYM4iapIqZ3r6966/ayySr
MA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgEGMAoGCCqGSM49BAMDA2gA
MGUCMQCD6cHEFl4aXTQY2e3v9GwOAEZLuN+yRhHFD/3meoyhpmvOwgPUnPWTxnS4
at+qIxUCMG1mihDK1A3UT82NQz60imOlM27jbdoXt2QfyFMm+YhidDkLF1vLUagM
6BgD56KyKA==
-----END CERTIFICATE-----`
	appStoreServerNotification, err := util.New(request.SignedPayload, rootCert)
	if err != nil {
		return nil, err
	}
	if !appStoreServerNotification.IsValid {
		err = fmt.Errorf("app store server notification invalid: %w", err)
		logger.Error("validate notification failed", zap.NamedError("appError", err))
		if err := daoManager.RawTransactionsDAO.UpdateHandleErrorByID(ctx, rawTransaction.ID, err.Error()); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
		}
		return nil, err
	}
	return appStoreServerNotification, nil
}

func checkAppAccountTokenInNotification(ctx context.Context, daoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction, appStoreServerNotification *util.AppStoreServerNotification) (*SlarkUser, error) {
	accountToken, err := daoManager.SlarkUserDAO.GetByAppAccountToken(
		ctx,
		appStoreServerNotification.TransactionInfo.AppAccountToken,
	)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	if accountToken == nil {
		err = fmt.Errorf("invalid app account token [%s]", appStoreServerNotification.TransactionInfo.AppAccountToken)
		logger.Warn("get app account token", zap.NamedError("appError", err))
		if err := daoManager.RawTransactionsDAO.UpdateHandleErrorByID(ctx, rawTransaction.ID, err.Error()); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
		}
		return nil, err
	}
	return accountToken, nil
}

func handleTransactionRecord(ctx context.Context, environment string, daoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction, accountToken *SlarkUser, appStoreServerNotification *util.AppStoreServerNotification) error {
	if rawTransaction.Environment == consts.DESIGNATED_ENVIRONMENT_NUM_PROD || (rawTransaction.Environment != consts.DESIGNATED_ENVIRONMENT_NUM_SANDBOX && environment == consts.DESIGNATED_ENVIRONMENT_PROD) {
		if err := insertTransactionsProd(
			ctx,
			daoManager,
			accountToken,
			appStoreServerNotification.TransactionInfo,
			appStoreServerNotification.RenewalInfo,
		); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
	} else {
		if err := insertTransactionsTest(
			ctx,
			daoManager,
			accountToken,
			appStoreServerNotification.TransactionInfo,
			appStoreServerNotification.RenewalInfo,
		); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
	}
	return nil
}

func handleStateRecordProd(ctx context.Context, tx *gorm.DB, daoManager, newDaoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction, accountToken *SlarkUser, appStoreServerNotification *util.AppStoreServerNotification) (bool, error) {
	var subscriptionState *SubscriptionStateProd
	var err error
	subscriptionState, err = newDaoManager.SubscriptionStateProdDAO.GetByUserID(ctx, accountToken.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return false, err
	}
	if subscriptionState == nil {
		if err = insertSubscriptionStateProd(
			ctx,
			newDaoManager,
			accountToken,
			appStoreServerNotification.Payload.NotificationType,
			appStoreServerNotification.TransactionInfo,
			appStoreServerNotification.RenewalInfo,
		); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return false, err
		}
		if err = tx.Commit().Error; err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return false, err
		}
		// delete handled raw_transactions record
		if err := daoManager.RawTransactionsDAO.DeleteHandleByID(ctx, rawTransaction.ID, ""); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return false, err
		}
		return true, nil
	}
	switch appStoreServerNotification.Payload.NotificationType {
	case consts.DID_CHANGE_RENEWAL_STATUS, consts.DID_FAIL_TO_RENEW, consts.GRACE_PERIOD_EXPIRED,
		consts.OFFER_REDEEMED, consts.REFUND_DECLINED, consts.REFUND_REVERSED:
	case consts.DID_RENEW, consts.SUBSCRIBED:
		subscriptionState.Subscribed = true
	case consts.EXPIRED, consts.REFUND:
		subscriptionState.Subscribed = false
	}
	subscriptionState.SubscribedAt = appStoreServerNotification.TransactionInfo.OriginalPurchaseDate
	subscriptionState.CurrentBillDate = appStoreServerNotification.TransactionInfo.PurchaseDate
	subscriptionState.CurrentBillPrice = decimal.New(int64(appStoreServerNotification.TransactionInfo.Price), -3).String()
	subscriptionState.CurrentOfferType = consts.OfferType(appStoreServerNotification.TransactionInfo.OfferType)
	subscriptionState.CurrentOfferID = appStoreServerNotification.TransactionInfo.OfferIdentifier
	subscriptionState.ExpirationDate = appStoreServerNotification.TransactionInfo.ExpiresDate
	subscriptionState.AutoRenew = appStoreServerNotification.RenewalInfo.AutoRenewStatus == 1
	subscriptionState.NextOfferType = consts.OfferType(appStoreServerNotification.RenewalInfo.OfferType)
	subscriptionState.NextOfferID = appStoreServerNotification.RenewalInfo.OfferIdentifier
	subscriptionState.CurrencyCode = appStoreServerNotification.TransactionInfo.Currency
	if err = newDaoManager.SubscriptionStateProdDAO.UpdateByID(ctx, subscriptionState.ID, subscriptionState); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return false, err
	}
	return false, nil
}

func handleStateRecordTest(ctx context.Context, tx *gorm.DB, daoManager, newDaoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction, accountToken *SlarkUser, appStoreServerNotification *util.AppStoreServerNotification) (bool, error) {
	var subscriptionState *SubscriptionStateTest
	var err error
	subscriptionState, err = newDaoManager.SubscriptionStateTestDAO.GetByUserID(ctx, accountToken.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return false, err
	}
	if subscriptionState == nil {
		if err = insertSubscriptionStateTest(
			ctx,
			newDaoManager,
			accountToken,
			appStoreServerNotification.Payload.NotificationType,
			appStoreServerNotification.TransactionInfo,
			appStoreServerNotification.RenewalInfo,
		); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return false, err
		}
		if err = tx.Commit().Error; err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return false, err
		}
		// delete handled raw_transactions record
		if err := daoManager.RawTransactionsDAO.DeleteHandleByID(ctx, rawTransaction.ID, ""); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return false, err
		}
		return true, nil
	}
	switch appStoreServerNotification.Payload.NotificationType {
	case consts.DID_CHANGE_RENEWAL_STATUS, consts.DID_FAIL_TO_RENEW, consts.GRACE_PERIOD_EXPIRED,
		consts.OFFER_REDEEMED, consts.REFUND_DECLINED, consts.REFUND_REVERSED:
	case consts.DID_RENEW, consts.SUBSCRIBED:
		subscriptionState.Subscribed = true
	case consts.EXPIRED, consts.REFUND:
		subscriptionState.Subscribed = false
	}
	subscriptionState.SubscribedAt = appStoreServerNotification.TransactionInfo.OriginalPurchaseDate
	subscriptionState.CurrentBillDate = appStoreServerNotification.TransactionInfo.PurchaseDate
	subscriptionState.CurrentBillPrice = decimal.New(int64(appStoreServerNotification.TransactionInfo.Price), -3).String()
	subscriptionState.CurrentOfferType = consts.OfferType(appStoreServerNotification.TransactionInfo.OfferType)
	subscriptionState.CurrentOfferID = appStoreServerNotification.TransactionInfo.OfferIdentifier
	subscriptionState.ExpirationDate = appStoreServerNotification.TransactionInfo.ExpiresDate
	subscriptionState.AutoRenew = appStoreServerNotification.RenewalInfo.AutoRenewStatus == 1
	subscriptionState.NextOfferType = consts.OfferType(appStoreServerNotification.RenewalInfo.OfferType)
	subscriptionState.NextOfferID = appStoreServerNotification.RenewalInfo.OfferIdentifier
	subscriptionState.CurrencyCode = appStoreServerNotification.TransactionInfo.Currency
	if err = newDaoManager.SubscriptionStateTestDAO.UpdateByID(ctx, subscriptionState.ID, subscriptionState); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return false, err
	}
	return false, nil
}

func handleDiscountOffer(ctx context.Context, environment string, tx *gorm.DB, daoManager, newDaoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction, accountToken *SlarkUser, appStoreServerNotification *util.AppStoreServerNotification) (bool, error) {
	var err error
	if appStoreServerNotification.TransactionInfo.OfferType == consts.OfferTypePromotional &&
		(appStoreServerNotification.TransactionInfo.OfferIdentifier == util.AppConfig(appStoreServerNotification.TransactionInfo.BundleId).DiscountOffer.IDNewUser ||
			appStoreServerNotification.TransactionInfo.OfferIdentifier == util.AppConfig(appStoreServerNotification.TransactionInfo.BundleId).DiscountOffer.ID10M ||
			appStoreServerNotification.TransactionInfo.OfferIdentifier == util.AppConfig(appStoreServerNotification.TransactionInfo.BundleId).DiscountOffer.ID8M ||
			appStoreServerNotification.TransactionInfo.OfferIdentifier == util.AppConfig(appStoreServerNotification.TransactionInfo.BundleId).DiscountOffer.ID6M ||
			appStoreServerNotification.TransactionInfo.OfferIdentifier == util.AppConfig(appStoreServerNotification.TransactionInfo.BundleId).DiscountOffer.ID4M ||
			appStoreServerNotification.TransactionInfo.OfferIdentifier == util.AppConfig(appStoreServerNotification.TransactionInfo.BundleId).DiscountOffer.ID2M) {
		var newUserDiscountState *NewUserDiscountState
		newUserDiscountState, err = newDaoManager.NewUserDiscountStateDAO.GetByUserID(ctx, accountToken.UserID)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				logger.Error("internal error", zap.NamedError("appError", err))
				return false, err
			} else {
				// logger.Info("no NewUserDiscountState log")
				if err = tx.Commit().Error; err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return false, err
				}
				// delete handled raw_transactions record
				if err := daoManager.RawTransactionsDAO.DeleteHandleByID(ctx, rawTransaction.ID, ""); err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return false, err
				}
				return true, nil
			}
		}
		if rawTransaction.Environment == consts.DESIGNATED_ENVIRONMENT_NUM_PROD || (rawTransaction.Environment != consts.DESIGNATED_ENVIRONMENT_NUM_SANDBOX && environment == consts.DESIGNATED_ENVIRONMENT_PROD) {
			newUserDiscountState.BilledTimes += 1
			newUserDiscountState.RemainingTimes -= 1
			if err = newDaoManager.NewUserDiscountStateDAO.UpdateByID(ctx, newUserDiscountState.ID, newUserDiscountState); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return false, err
			}
		} else {
			logger.Warn("sandbox environment no need to handle offer")
			if err = tx.Commit().Error; err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return false, err
			}
			// delete handled raw_transactions record
			if err := daoManager.RawTransactionsDAO.DeleteHandleByID(ctx, rawTransaction.ID, ""); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func handleTransaction(ctx context.Context, environment string, daoManager *dao.Manager, logger *zap.Logger, rawTransaction *RawTransaction, accountToken *SlarkUser, appStoreServerNotification *util.AppStoreServerNotification) (bool, error) {
	var err error
	tx, newDaoManager := daoManager.Transaction()
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback().Error; rbErr != nil {
				logger.Error("tx.Rollback().Error", zap.NamedError("appError", rbErr))
			} else {
				logger.Info("tx.Rollback() success")
			}
		}
	}()
	// handle state record
	if rawTransaction.Environment == consts.DESIGNATED_ENVIRONMENT_NUM_PROD || (rawTransaction.Environment != consts.DESIGNATED_ENVIRONMENT_NUM_SANDBOX && environment == consts.DESIGNATED_ENVIRONMENT_PROD) {
		needContinue, err := handleStateRecordProd(ctx, tx, daoManager, newDaoManager, logger, rawTransaction, accountToken, appStoreServerNotification)
		if err != nil {
			return false, err
		}
		if needContinue {
			return needContinue, nil
		}
	} else {
		needContinue, err := handleStateRecordTest(ctx, tx, daoManager, newDaoManager, logger, rawTransaction, accountToken, appStoreServerNotification)
		if err != nil {
			return false, err
		}
		if needContinue {
			return needContinue, nil
		}
	}
	// handler offer
	if appStoreServerNotification.Payload.NotificationType == consts.DID_RENEW {
		needContinue, err := handleDiscountOffer(ctx, environment, tx, daoManager, newDaoManager, logger, rawTransaction, accountToken, appStoreServerNotification)
		if err != nil {
			return false, err
		}
		if needContinue {
			return needContinue, nil
		}
	}
	if err = tx.Commit().Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return false, err
	}
	// delete handled raw_transactions record
	if err := daoManager.RawTransactionsDAO.DeleteHandleByID(ctx, rawTransaction.ID, ""); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return false, err
	}
	return false, nil
}

func handleAppStoreNotification(daoManager *dao.Manager, logger *zap.Logger) error {
	ctx := context.Background()
	// get unhandled notification from raw_transactions table
	rows, err := daoManager.RawTransactionsDAO.GetUnhandledRows(ctx)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var err error
		rawTransaction := &RawTransaction{}
		if err = daoManager.DB.ScanRows(rows, rawTransaction); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
		// parse appstore server notification
		appStoreServerNotification, err := parseAppStoreServerNotification(ctx, daoManager, logger, rawTransaction)
		if err != nil {
			continue
		}
		// check app account token
		accountToken, err := checkAppAccountTokenInNotification(ctx, daoManager, logger, rawTransaction, appStoreServerNotification)
		if err != nil {
			continue
		}
		// get environment key
		environment := strings.ToLower(strings.TrimSpace(appStoreServerNotification.TransactionInfo.Environment))
		// handle transaction record
		switch appStoreServerNotification.Payload.NotificationType {
		case consts.DID_CHANGE_RENEWAL_STATUS, consts.DID_FAIL_TO_RENEW,
			consts.DID_RENEW, consts.EXPIRED, consts.GRACE_PERIOD_EXPIRED,
			consts.OFFER_REDEEMED, consts.REFUND, consts.REFUND_DECLINED,
			consts.REFUND_REVERSED, consts.SUBSCRIBED:
			if err = handleTransactionRecord(ctx, environment, daoManager, logger, rawTransaction, accountToken, appStoreServerNotification); err != nil {
				return err
			}
		case consts.TEST:
			err = fmt.Errorf("notification type: %v", appStoreServerNotification.Payload.NotificationType)
			if err := daoManager.RawTransactionsDAO.DeleteHandleByID(ctx, rawTransaction.ID, err.Error()); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return err
			}
			continue
		default:
			err = fmt.Errorf("unknown notification type: %v", appStoreServerNotification.Payload.NotificationType)
			if err := daoManager.RawTransactionsDAO.UpdateHandleErrorByID(ctx, rawTransaction.ID, err.Error()); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
			}
			continue
		}
		// handle transaction
		needContinue, err := handleTransaction(ctx, environment, daoManager, logger, rawTransaction, accountToken, appStoreServerNotification)
		if err != nil {
			return err
		}
		if needContinue {
			continue
		}
	}
	if err := rows.Err(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	logger.Info("handle appstore notifications, one turn over")
	return nil
}
