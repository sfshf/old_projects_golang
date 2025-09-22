package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	"github.com/sethvargo/go-password/password"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PrivacyEmailService struct {
	*PswdsService
}

func NewPrivacyEmailService(ctx context.Context, pswdsService *PswdsService) (*PrivacyEmailService, error) {
	s := &PrivacyEmailService{
		PswdsService: pswdsService,
	}
	return s, nil
}

func (s *PrivacyEmailService) handleEmail(emailAccount, mailbox string, fetchCmd *imapclient.FetchCommand) ([]*PrivacyEmail, error) {
	now := time.Now()
	var newEmails []*PrivacyEmail
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}
		var bodySection imapclient.FetchItemDataBodySection
		ok := false
		for {
			item := msg.Next()
			if item == nil {
				break
			}
			bodySection, ok = item.(imapclient.FetchItemDataBodySection)
			if ok {
				break
			}
		}
		if !ok {
			break
		}
		mr, err := mail.CreateReader(bodySection.Literal)
		if err != nil {
			return nil, err
		}
		h := mr.Header
		sentAt, err := h.Date()
		if err != nil {
			return nil, err
		}
		if sentAt.Before(now.Add(-15 * 24 * time.Hour)) {
			continue
		}
		var sentBy string
		from, err := h.AddressList("From")
		if err != nil {
			return nil, err
		}
		if len(from) > 0 {
			var sb strings.Builder
			for idx, item := range from {
				name := strings.TrimSpace(item.Name)
				if name != "" {
					sb.WriteString(name)
					sb.WriteString(" (")
				}
				sb.WriteString(item.Address)
				if name != "" {
					sb.WriteString(")")
				}
				if idx != len(from)-1 {
					sb.WriteRune('\n')
				}
			}
			sentBy = sb.String()
		}
		subject, err := h.Text("Subject")
		if err != nil {
			return nil, err
		}
		newEmails = append(newEmails, &PrivacyEmail{
			EmailAccount: emailAccount,
			Mailbox:      mailbox,
			UID:          int64(msg.SeqNum),
			SentAt:       sentAt.UnixMilli(),
			SentBy:       sentBy,
			Subject:      base64.StdEncoding.EncodeToString([]byte(subject)),
		})
	}
	if err := fetchCmd.Close(); err != nil {
		return nil, err
	}
	return newEmails, nil
}

func (s *PrivacyEmailService) cacheAllHeaders(ctx context.Context, mailCli *imapclient.Client, emailAccount string, last uint32) error {
	var seqSet imap.SeqSet
	seqSet.AddRange(1, last)
	fetchOptions := &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{{Specifier: imap.PartSpecifierHeader}},
	}
	fetchCmd := mailCli.Fetch(seqSet, fetchOptions)
	newEmails, err := s.handleEmail(emailAccount, mailCli.Mailbox().Name, fetchCmd)
	if err != nil {
		return err
	}
	if len(newEmails) > 0 {
		if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
			daoManager := dao.ManagerWithDB(tx)
			if err := daoManager.PrivacyEmailDAO.DeleteByMailbox(ctx, emailAccount, mailCli.Mailbox().Name); err != nil {
				return err
			}
			if err := daoManager.PrivacyEmailContentDAO.DeleteByMailbox(ctx, emailAccount, mailCli.Mailbox().Name); err != nil {
				return err
			}
			if err := daoManager.PrivacyEmailDAO.Create(ctx, newEmails); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *PrivacyEmailService) checkFirstAndLast(ctx context.Context, mailCli *imapclient.Client, emailAccount string, last uint32) (bool, error) {
	allCached := true
	var seqSet imap.SeqSet
	seqSet.AddNum(1)
	if last > 1 {
		seqSet.AddNum(last)
	}
	fetchOptions := &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{{Specifier: imap.PartSpecifierHeader}},
	}
	fetchCmd := mailCli.Fetch(seqSet, fetchOptions)
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}
		var bodySection imapclient.FetchItemDataBodySection
		ok := false
		for {
			item := msg.Next()
			if item == nil {
				break
			}
			bodySection, ok = item.(imapclient.FetchItemDataBodySection)
			if ok {
				break
			}
		}
		if allCached = allCached && ok; !allCached {
			break
		}
		// Read the message via the go-message library
		mr, err := mail.CreateReader(bodySection.Literal)
		if err != nil {
			return false, err
		}
		// a few header fields
		h := mr.Header
		sentAt, err := h.Date()
		if err != nil {
			return false, err
		}
		var sentBy string
		from, err := h.AddressList("From")
		if err != nil {
			return false, err
		}
		if len(from) > 0 {
			var sb strings.Builder
			for idx, item := range from {
				name := strings.TrimSpace(item.Name)
				if name != "" {
					sb.WriteString(name)
					sb.WriteString(" (")
				}
				sb.WriteString(item.Address)
				if name != "" {
					sb.WriteString(")")
				}
				if idx != len(from)-1 {
					sb.WriteRune('\n')
				}
			}
			sentBy = sb.String()
		}
		subject, err := h.Text("Subject")
		if err != nil {
			return false, err
		}
		// check the email cache
		record, err := s.DaoManager.PrivacyEmailDAO.GetByHeaders(ctx, emailAccount, mailCli.Mailbox().Name, msg.SeqNum, sentAt.UnixMilli(), sentBy, subject)
		if err != nil {
			return false, err
		}
		if allCached = allCached && record != nil; !allCached {
			break
		}
	}
	if err := fetchCmd.Close(); err != nil {
		return false, err
	}
	return allCached, nil
}

func (s *PrivacyEmailService) checkEveryMailbox(ctx context.Context, mailCli *imapclient.Client, emailAccount string) error {
	mailboxes := []string{"INBOX", "Spam", "Trash", "Archive"}
	for _, mailbox := range mailboxes {
		selectedMbox, err := mailCli.Select(mailbox, nil).Wait()
		if err != nil {
			return err
		}
		if selectedMbox.NumMessages == 0 {
			// clear all cached emails
			emails, err := s.DaoManager.PrivacyEmailDAO.GetByMailbox(ctx, emailAccount, mailCli.Mailbox().Name)
			if err != nil {
				return err
			}
			if len(emails) > 0 {
				if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
					daoManager := dao.ManagerWithDB(tx)
					if err := daoManager.PrivacyEmailDAO.DeleteByMailbox(ctx, emailAccount, mailCli.Mailbox().Name); err != nil {
						return err
					}
					for _, item := range emails {
						if err := daoManager.PrivacyEmailContentDAO.DeleteByEmailID(ctx, item.ID); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
			}
			continue
		}
		// 1. check the first and the last emails
		allCached, err := s.checkFirstAndLast(ctx, mailCli, emailAccount, selectedMbox.NumMessages)
		if err != nil {
			return err
		}
		if allCached {
			continue
		}
		// 2. cache all (headers) if not all cached
		if err := s.cacheAllHeaders(ctx, mailCli, emailAccount, selectedMbox.NumMessages); err != nil {
			return err
		}
	}
	return nil
}

func (s *PrivacyEmailService) checkEveryEmailAccount(ctx context.Context, accounts []*PrivacyEmailAccount) error {
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		return err
	}
	defer func() {
		cli.Close()
	}()
	for _, account := range accounts {
		if err := cli.Login(account.EmailAccount, account.Password).Wait(); err != nil {
			return err
		}
		// check every mail box
		if err := s.checkEveryMailbox(ctx, cli, account.EmailAccount); err != nil {
			return err
		}
		if err := cli.Logout().Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (s *PrivacyEmailService) getPrivacyEmails(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetPrivacyEmailsRequest, accounts []*PrivacyEmailAccount) (*pswds_api.GetPrivacyEmailsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var emailAccounts []string
	for _, item := range accounts {
		emailAccounts = append(emailAccounts, item.EmailAccount)
	}
	tx := s.DaoManager.PrivacyEmailDAO.Table(ctx).
		Where(`email_account IN (?)`, emailAccounts).
		Order(`sent_at DESC`)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if total == 0 {
		return &pswds_api.GetPrivacyEmailsResponse_Data{
			HasAccount: true,
			Accounts:   emailAccounts,
		}, nil
	}
	var emails []*PrivacyEmail
	if err := tx.Offset(int(req.PageNum * req.PageSize)).
		Limit(int(req.PageSize)).
		Find(&emails).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*pswds_api.GetPrivacyEmailsResponse_Email
	for _, email := range emails {
		list = append(list, &pswds_api.GetPrivacyEmailsResponse_Email{
			Id:      email.ID,
			Mailbox: email.Mailbox,
			Uid:     email.UID,
			SentBy:  email.SentBy,
			SentAt:  email.SentAt,
			Subject: email.Subject,
		})
	}
	return &pswds_api.GetPrivacyEmailsResponse_Data{
		HasAccount: true,
		Total:      total,
		List:       list,
		Accounts:   emailAccounts,
	}, nil
}

func (s *PrivacyEmailService) GetPrivacyEmails(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetPrivacyEmailsRequest) (*pswds_api.GetPrivacyEmailsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// 1. check user's privacy email account
	accounts, err := s.DaoManager.PrivacyEmailAccountDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if len(accounts) == 0 {
		return &pswds_api.GetPrivacyEmailsResponse_Data{
			HasAccount: false,
		}, nil
	}
	if err := s.checkEveryEmailAccount(ctx, accounts); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 2. get list from db
	if req.PageSize > 20 {
		req.PageSize = 20
	}
	return s.getPrivacyEmails(ctx, rpcCtx, req, accounts)
}

func detectImageType(data []byte) (string, bool) {
	// JPEG
	// 	- start with(2 bytes): ff d8
	//	- end with(2 bytes): ff d9
	jpegStart, _ := hex.DecodeString("ffd8")
	jpegEnd, _ := hex.DecodeString("ffd9")
	if bytes.HasPrefix(data, jpegStart) && bytes.HasSuffix(data, jpegEnd) {
		return "JPEG", true
	}
	// PNG
	// 	- start with(8 bytes): 89 50 4E 47 0D 0A 1A 0A
	pngStart, _ := hex.DecodeString("89504E470D0A1A0A")
	if bytes.HasPrefix(data, pngStart) {
		return "PNG", true
	}
	// GIF
	//	- start with(6 bytes): 47 49 46 38 39/37 61
	gifStart1, _ := hex.DecodeString("474946383961")
	gifStart2, _ := hex.DecodeString("474946383761")
	if bytes.HasPrefix(data, gifStart1) {
		return "GIF", true
	} else if bytes.HasPrefix(data, gifStart2) {
		return "GIF", true
	}
	// TIFF
	//	- start with(2 bytes): 4D 4D or 49 49
	tiffStart1, _ := hex.DecodeString("4D4D")
	tiffStart2, _ := hex.DecodeString("4949")
	if bytes.HasPrefix(data, tiffStart1) {
		return "TIFF", true
	} else if bytes.HasPrefix(data, tiffStart2) {
		return "TIFF", true
	}
	// ICO
	//	- start with(8 bytes): 00 00 01 00 01 00 20 20
	icoStart, _ := hex.DecodeString("0000010001002020")
	if bytes.HasPrefix(data, icoStart) {
		return "ICO", true
	}
	// IFF
	// 	- start with(4 bytes): 46 4F 52 4D
	iffStart, _ := hex.DecodeString("464F524D")
	if bytes.HasPrefix(data, iffStart) {
		return "IFF", true
	}
	return "", false
}

func base64ImagePrefix(imgType string) string {
	return "image/" + strings.ToLower(imgType) + ","
}

func (s *PrivacyEmailService) validatePrivacyEmail(ctx context.Context, rpcCtx *rpc.Context, id int64, userID int64) (*PrivacyEmail, *PrivacyEmailAccount, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// 1. get headers from db
	record, err := s.DaoManager.PrivacyEmailDAO.GetByID(ctx, id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record == nil {
		err = fmt.Errorf("invalid id parameter")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// 2. get email account
	emailAccount, err := s.DaoManager.PrivacyEmailAccountDAO.GetByUserIDAndAccount(ctx, userID, record.EmailAccount)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if emailAccount == nil {
		err = errors.New("privacy email account not found")
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return record, emailAccount, nil
}

func (s *PrivacyEmailService) validateEmailHeaders(rpcCtx *rpc.Context, record *PrivacyEmail, mr *mail.Reader) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	h := mr.Header
	var sentAt int64 // sentAt
	if date, err := h.Date(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	} else {
		sentAt = date.UnixMilli()
	}
	var sentBy string // sentBy
	if from, err := h.AddressList("From"); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	} else {
		if len(from) > 0 {
			var sb strings.Builder
			for idx, item := range from {
				name := strings.TrimSpace(item.Name)
				if name != "" {
					sb.WriteString(name)
					sb.WriteString(" (")
				}
				sb.WriteString(item.Address)
				if name != "" {
					sb.WriteString(")")
				}
				if idx != len(from)-1 {
					sb.WriteRune('\n')
				}
			}
			sentBy = sb.String()
		}
	}
	subject, err := h.Text("Subject") // subject
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// check headers
	if record.SentAt != sentAt || record.SentBy != sentBy || record.Subject != base64.StdEncoding.EncodeToString([]byte(subject)) {
		err = errors.New("inconsistent cache data")
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *PrivacyEmailService) handleEmailBody(rpcCtx *rpc.Context, record *PrivacyEmail, p *mail.Part) (*PrivacyEmailContent, *pswds_api.GetPrivacyEmailResponse_Content, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	b, err := io.ReadAll(p.Body)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	content := base64.StdEncoding.EncodeToString(b)
	body := &pswds_api.GetPrivacyEmailResponse_Content{
		ContentType: "text",
		Content:     content,
	}
	one := &PrivacyEmailContent{
		EmailID:     record.ID,
		Type:        "body",
		ContentType: "text",
		Filesize:    int64(len(b)),
		Content:     content,
	}
	if imageType, is := detectImageType(b); is {
		one.ContentType = "image"
		one.Content = "data:" + base64ImagePrefix(imageType) + "base64," + content
		body.ContentType = "image"
		body.Content = "data:" + base64ImagePrefix(imageType) + "base64," + content
	}
	return one, body, nil
}

func (s *PrivacyEmailService) handleEmailAttachments(rpcCtx *rpc.Context, record *PrivacyEmail, p *mail.Part) (*PrivacyEmailContent, *pswds_api.GetPrivacyEmailResponse_Attachment, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// check image file
	contentType := p.Header.(*mail.AttachmentHeader).Get("Content-Type")
	b, err := io.ReadAll(p.Body)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	contentDisposition := p.Header.(*mail.AttachmentHeader).Get("Content-Disposition")
	filename := strings.TrimSuffix(strings.TrimPrefix(contentDisposition, `attachment; filename="`), `"`)
	attachment := &pswds_api.GetPrivacyEmailResponse_Attachment{
		Filename: filename,
		Size:     int64(len(b)),
	}
	one := &PrivacyEmailContent{
		EmailID:     record.ID,
		Type:        "attachment",
		ContentType: "bytes",
		Filename:    filename,
		Filesize:    int64(len(b)),
	}
	if strings.Contains(contentType, "image") {
		one.ContentType = "image"
		splits := strings.Split(contentType, " ")
		if len(filename) == 0 {
			if len(splits) == 2 {
				if strings.HasPrefix(splits[1], `name="`) {
					one.Filename = strings.TrimSuffix(strings.TrimPrefix(splits[1], `name="`), `"`)
				}
			}
		}
		content := "data:" + strings.TrimSpace(splits[0]) + "base64," + base64.StdEncoding.EncodeToString(b)
		one.Content = content
		attachment.Content = content // set attachment content only when it is an image
	} else {
		one.Content = base64.StdEncoding.EncodeToString(b)
	}
	return one, attachment, nil
}

func (s *PrivacyEmailService) handleEmailContents(ctx context.Context, rpcCtx *rpc.Context, data *pswds_api.GetPrivacyEmailResponse_Data, record *PrivacyEmail, mr *mail.Reader) (*pswds_api.GetPrivacyEmailResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// contents
	var contents []*PrivacyEmailContent
	var bodys []*pswds_api.GetPrivacyEmailResponse_Content
	var attachments []*pswds_api.GetPrivacyEmailResponse_Attachment
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		switch p.Header.(type) {
		case *mail.InlineHeader: // email body
			content, body, appError := s.handleEmailBody(rpcCtx, record, p)
			if appError != nil {
				return nil, appError
			}
			contents = append(contents, content)
			bodys = append(bodys, body)
		case *mail.AttachmentHeader: // email attachments
			content, attachment, appError := s.handleEmailAttachments(rpcCtx, record, p)
			if appError != nil {
				return nil, appError
			}
			contents = append(contents, content)
			attachments = append(attachments, attachment)
		}
	}
	data.Contents = bodys
	data.Attachments = attachments
	// cache the contents of the email
	if len(contents) > 0 {
		if err := s.DaoManager.PrivacyEmailContentDAO.Create(ctx, contents); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return data, nil
}

func (s *PrivacyEmailService) getRemotePrivacyEmail(ctx context.Context, rpcCtx *rpc.Context, data *pswds_api.GetPrivacyEmailResponse_Data, record *PrivacyEmail, emailAccount *PrivacyEmailAccount) (*pswds_api.GetPrivacyEmailResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer func() {
		cli.Close()
	}()
	if err := cli.Login(emailAccount.EmailAccount, emailAccount.Password).Wait(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer func() {
		cli.Logout().Wait()
	}()
	if _, err = cli.Select(record.Mailbox, nil).Wait(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	seqSet := imap.SeqSetNum(uint32(record.UID))
	fetchOptions := &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{{}},
	}
	fetchCmd := cli.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()
	msg := fetchCmd.Next()
	if msg == nil {
		return data, nil
	}
	var bodySection imapclient.FetchItemDataBodySection
	ok := false
	for {
		item := msg.Next()
		if item == nil {
			break
		}
		bodySection, ok = item.(imapclient.FetchItemDataBodySection)
		if ok {
			break
		}
	}
	if !ok {
		err = errors.New("parse email message body fail")
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	mr, err := mail.CreateReader(bodySection.Literal)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// validate email headers
	if appError := s.validateEmailHeaders(rpcCtx, record, mr); appError != nil {
		return nil, appError
	}
	// handle email contents
	return s.handleEmailContents(ctx, rpcCtx, data, record, mr)
}

func (s *PrivacyEmailService) GetPrivacyEmail(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetPrivacyEmailRequest) (*pswds_api.GetPrivacyEmailResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	record, emailAccount, appError := s.validatePrivacyEmail(ctx, rpcCtx, req.Id, loginInfo.UserID)
	if appError != nil {
		return nil, appError
	}
	data := &pswds_api.GetPrivacyEmailResponse_Data{
		Id:      record.ID,
		Mailbox: record.Mailbox,
		Uid:     record.UID,
		SentBy:  record.SentBy,
		SentAt:  record.SentAt,
		Subject: record.Subject,
	}
	// 3. try to get contents from db
	emailContents, err := s.DaoManager.PrivacyEmailContentDAO.GetByEmailID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if len(emailContents) > 0 {
		var bodys []*pswds_api.GetPrivacyEmailResponse_Content
		var attachments []*pswds_api.GetPrivacyEmailResponse_Attachment
		for _, item := range emailContents {
			if item.Type == "body" { // body|attachment
				bodys = append(bodys, &pswds_api.GetPrivacyEmailResponse_Content{
					ContentType: item.ContentType, // text|image|bytes
					Content:     item.Content,
				})
			} else if item.Type == "attachment" {
				one := &pswds_api.GetPrivacyEmailResponse_Attachment{
					Filename: item.Filename,
					Size:     item.Filesize,
				}
				if item.ContentType == "image" { // text|image|bytes
					one.Content = item.Content
				}
				attachments = append(attachments, one)
			}
		}
		data.Contents = bodys
		data.Attachments = attachments
		return data, nil
	}
	// 4. fetch from remote, if has no content in db.
	return s.getRemotePrivacyEmail(ctx, rpcCtx, data, record, emailAccount)
}

func (s *PrivacyEmailService) deleteRemoteEmail(rpcCtx *rpc.Context, record *PrivacyEmail, emailAccount *PrivacyEmailAccount) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer func() {
		cli.Close()
	}()
	if err := cli.Login(emailAccount.EmailAccount, emailAccount.Password).Wait(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer func() {
		cli.Logout().Wait()
	}()
	// 3. remove the remote email
	if _, err = cli.Select(record.Mailbox, nil).Wait(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	seqSet := imap.SeqSetNum(uint32(record.UID))
	storeFlags := imap.StoreFlags{
		Op:    imap.StoreFlagsAdd,
		Flags: []imap.Flag{imap.FlagDeleted},
	}
	if err := cli.Store(seqSet, &storeFlags, nil).Close(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if _, err = cli.Expunge().Collect(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *PrivacyEmailService) DeletePrivacyEmail(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.DeletePrivacyEmailRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate privacy emial record
	record, emailAccount, appError := s.validatePrivacyEmail(ctx, rpcCtx, req.Id, loginInfo.UserID)
	if appError != nil {
		return appError
	}
	// remove the remote email
	if appError := s.deleteRemoteEmail(rpcCtx, record, emailAccount); appError != nil {
		return appError
	}
	// delete data in db, and reorder data
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// delete the email
		if err := daoManager.PrivacyEmailDAO.DeleteByID(ctx, record.ID); err != nil {
			return err
		}
		if err := daoManager.PrivacyEmailContentDAO.DeleteByEmailID(ctx, record.ID); err != nil {
			return err
		}
		// reorder emails what's uid greater than the email's
		if err := daoManager.PrivacyEmailDAO.ReorderUids(ctx, emailAccount.EmailAccount, record.Mailbox, record.UID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func addPrivacyEmailAccount(account, password string) error {
	account = strings.ToLower(account)
	data := url.Values{}
	data.Add("email", account)
	data.Add("password", password)
	data.Add("privileges", "")
	req, err := http.NewRequest(http.MethodPost, "https://box.n1xt.net/admin/mail/users/add", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(os.Getenv("PRIVACY_EMAIL_ADMIN_ACCOUNT") + ":" + os.Getenv("PRIVACY_EMAIL_ADMIN_PASSWORD")))
	req.Header.Set("Authorization", "Basic "+auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("add privacy email account fail: " + resp.Status)
	}
	return nil
}

func (s *PrivacyEmailService) AddPrivacyEmailAccount(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.AddPrivacyEmailAccountRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// 1. add an remote email account
	account := strings.ToLower(random.NewUUIDHexEncoding()[:16] + "@n1xt.net")
	emailAccount, err := s.DaoManager.PrivacyEmailAccountDAO.GetByAccount(ctx, account)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if emailAccount != nil {
		account = strings.ToLower(random.NewUUIDHexEncoding()[:16] + "@n1xt.net")
	}
	passwd, err := password.Generate(16, 5, 5, false, false)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err := addPrivacyEmailAccount(account, passwd); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 2. insert into db
	if err := s.DaoManager.PrivacyEmailAccountDAO.Create(ctx, &PrivacyEmailAccount{
		UserID:       loginInfo.UserID,
		EmailAccount: account,
		Password:     passwd,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
