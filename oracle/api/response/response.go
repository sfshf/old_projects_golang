package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/nextsurfer/ground/pkg/localize"
)

type Response struct {
	Code         int           `json:"code"`
	Message      string        `json:"message"`
	DebugMessage string        `json:"debugMessage"`
	Data         interface{}   `json:"data"`
	Oracle       string        `json:"oracle,omitempty"`
	Req          *http.Request `json:"-"`
}

func DeferWriteResponse(ctx context.Context, w http.ResponseWriter, resp *Response) {
	var respData []byte
	if resp == nil || resp.Message == "" {
		resp = &Response{
			Code:         StatusCodeInternalServerError,
			Message:      ctx.Value(_localizerCtx).(*localize.Localizer).Localize("FatalErrMsg"),
			DebugMessage: "nil response struct pointer",
		}
	}
	respData, _ = json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

var (
	_localizerCtx struct{}
)

func WithLocalizer(r *http.Request, manager *localize.Manager) *http.Request {
	lang := r.Header.Get("Accept-Language")
	if lang == "" {
		lang = "en"
	}
	return r.WithContext(context.WithValue(r.Context(), _localizerCtx, manager.Localizer(lang)))
}

func MustMethodPost(r *http.Request, resp *Response) bool {
	if r.Method != http.MethodPost {
		if resp == nil {
			resp = &Response{}
		}
		resp.Code = StatusCodeMethodNotAllowed
		resp.Message = r.Context().Value(_localizerCtx).(*localize.Localizer).Localize("ClientErrMethodNotAllowed")
		resp.DebugMessage = "request method not allowed"
		return false
	}
	return true
}

func LocalizeMessage(ctx context.Context, id string) string {
	return ctx.Value(_localizerCtx).(*localize.Localizer).Localize(id)
}
