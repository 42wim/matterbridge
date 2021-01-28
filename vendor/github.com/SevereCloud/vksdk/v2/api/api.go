/*
Package api implements VK API.

See more https://vk.com/dev/api_requests
*/
package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SevereCloud/vksdk/v2"
	"github.com/SevereCloud/vksdk/v2/internal"
	"github.com/SevereCloud/vksdk/v2/object"
)

// Api constants.
const (
	Version   = vksdk.API
	MethodURL = "https://api.vk.com/method/"
)

// VKontakte API methods (except for methods from secure and ads sections)
// with user access key or service access key can be accessed
// no more than 3 times per second. The community access key is limited
// to 20 requests per second.
//
// Maximum amount of calls to the secure section methods depends
// on the app's users amount. If an app has less than
// 10 000 users, 5 requests per second,
// up to 100 000 – 8 requests,
// up to 1 000 000 – 20 requests,
// 1 000 000+ – 35 requests.
//
// The ads section methods are subject to their own limitations,
// you can read them on this page - https://vk.com/dev/ads_limits
//
// If one of this limits is exceeded, the server will return following error:
// "Too many requests per second". (errors.TooMany).
//
// If your app's logic implies many requests in a row, check the execute method.
// It allows for up to 25 requests for different methods in a single request.
//
// In addition to restrictions on the frequency of calls, there are also
// quantitative restrictions on calling the same type of methods.
//
// After exceeding the quantitative limit, access to a particular method may
// require entering a captcha (see https://vk.com/dev/captcha_error),
// and may also be temporarily restricted (in this case, the server does
// not return a response to the call of a particular method, but handles
// any other requests without problems).
//
// If this error occurs, the following parameters are also passed in
// the error message:
//
// CaptchaSID - identifier captcha.
//
// CaptchaImg - a link to the image that you want to show the user
// to enter text from that image.
//
// In this case, you should ask the user to enter text from
// the CaptchaImg image and repeat the request by adding parameters to it:
//
// captcha_sid - the obtained identifier;
//
// captcha_key - text entered by the user.
//
// More info: https://vk.com/dev/api_requests
const (
	LimitUserToken  = 3
	LimitGroupToken = 20
)

// VK struct.
type VK struct {
	accessTokens []string
	lastToken    uint32
	MethodURL    string
	Version      string
	Client       *http.Client
	Limit        int
	UserAgent    string
	Handler      func(method string, params ...Params) (Response, error)

	mux      sync.Mutex
	lastTime time.Time
	rps      int
}

// Response struct.
type Response struct {
	Response      json.RawMessage `json:"response"`
	Error         Error           `json:"error"`
	ExecuteErrors ExecuteErrors   `json:"execute_errors"`
}

// NewVK returns a new VK.
//
// The VKSDK will use the http.DefaultClient.
// This means that if the http.DefaultClient is modified by other components
// of your application the modifications will be picked up by the SDK as well.
//
// In some cases this might be intended, but it is a better practice
// to create a custom HTTP Client to share explicitly through
// your application. You can configure the VKSDK to use the custom
// HTTP Client by setting the VK.Client value.
//
// This set limit 20 requests per second for one token.
func NewVK(tokens ...string) *VK {
	var vk VK

	vk.accessTokens = tokens
	vk.Version = Version

	vk.Handler = vk.defaultHandler

	vk.MethodURL = MethodURL
	vk.Client = http.DefaultClient
	vk.Limit = LimitGroupToken
	vk.UserAgent = internal.UserAgent

	return &vk
}

// getToken return next token (simple round-robin).
func (vk *VK) getToken() string {
	i := atomic.AddUint32(&vk.lastToken, 1)
	return vk.accessTokens[(int(i)-1)%len(vk.accessTokens)]
}

// Params type.
type Params map[string]interface{}

// Lang - determines the language for the data to be displayed on. For
// example country and city names. If you use a non-cyrillic language,
// cyrillic symbols will be transliterated automatically.
// Numeric format from account.getInfo is supported as well.
//
// 	p.Lang(object.LangRU)
//
// See all language code in module object.
func (p Params) Lang(v int) Params {
	p["lang"] = v
	return p
}

// TestMode allows to send requests from a native app without switching it on
// for all users.
func (p Params) TestMode(v bool) Params {
	p["test_mode"] = v
	return p
}

// CaptchaSID received ID.
//
// See https://vk.com/dev/captcha_error
func (p Params) CaptchaSID(v string) Params {
	p["captcha_sid"] = v
	return p
}

// CaptchaKey text input.
//
// See https://vk.com/dev/captcha_error
func (p Params) CaptchaKey(v string) Params {
	p["captcha_key"] = v
	return p
}

// Confirm parameter.
//
// See https://vk.com/dev/need_confirmation
func (p Params) Confirm(v bool) Params {
	p["confirm"] = v
	return p
}

// WithContext parameter.
func (p Params) WithContext(ctx context.Context) Params {
	p[":context"] = ctx
	return p
}

func buildQuery(sliceParams ...Params) (context.Context, url.Values) {
	query := url.Values{}
	ctx := context.Background()

	for _, params := range sliceParams {
		for key, value := range params {
			if key != ":context" {
				query.Set(key, FmtValue(value, 0))
			} else {
				ctx = value.(context.Context)
			}
		}
	}

	return ctx, query
}

// defaultHandler provides access to VK API methods.
func (vk *VK) defaultHandler(method string, sliceParams ...Params) (Response, error) {
	u := vk.MethodURL + method
	ctx, query := buildQuery(sliceParams...)
	attempt := 0

	for {
		var response Response

		attempt++

		// Rate limiting
		if vk.Limit > 0 {
			vk.mux.Lock()

			sleepTime := time.Second - time.Since(vk.lastTime)
			if sleepTime < 0 {
				vk.lastTime = time.Now()
				vk.rps = 0
			} else if vk.rps == vk.Limit*len(vk.accessTokens) {
				time.Sleep(sleepTime)
				vk.lastTime = time.Now()
				vk.rps = 0
			}
			vk.rps++

			vk.mux.Unlock()
		}

		rawBody := bytes.NewBufferString(query.Encode())

		req, err := http.NewRequestWithContext(ctx, "POST", u, rawBody)
		if err != nil {
			return response, err
		}

		req.Header.Set("User-Agent", vk.UserAgent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := vk.Client.Do(req)
		if err != nil {
			return response, err
		}

		mediatype, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if mediatype != "application/json" {
			_ = resp.Body.Close()
			return response, &InvalidContentType{mediatype}
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			_ = resp.Body.Close()
			return response, err
		}

		_ = resp.Body.Close()

		switch response.Error.Code {
		case ErrNoType:
			return response, nil
		case ErrTooMany:
			if attempt < vk.Limit {
				continue
			}

			return response, &response.Error
		}

		return response, &response.Error
	}
}

// Request provides access to VK API methods.
func (vk *VK) Request(method string, sliceParams ...Params) ([]byte, error) {
	token := vk.getToken()

	reqParams := Params{
		"access_token": token,
		"v":            vk.Version,
	}

	sliceParams = append(sliceParams, reqParams)

	resp, err := vk.Handler(method, sliceParams...)

	return resp.Response, err
}

// RequestUnmarshal provides access to VK API methods.
func (vk *VK) RequestUnmarshal(method string, obj interface{}, sliceParams ...Params) error {
	rawResponse, err := vk.Request(method, sliceParams...)
	if err != nil {
		return err
	}

	return json.Unmarshal(rawResponse, &obj)
}

func fmtReflectValue(value reflect.Value, depth int) string {
	switch f := value; value.Kind() {
	case reflect.Invalid:
		return ""
	case reflect.Bool:
		return fmtBool(f.Bool())
	case reflect.Array, reflect.Slice:
		s := ""

		for i := 0; i < f.Len(); i++ {
			if i > 0 {
				s += ","
			}

			s += FmtValue(f.Index(i).Interface(), depth)
		}

		return s
	case reflect.Ptr:
		// pointer to array or slice or struct? ok at top level
		// but not embedded (avoid loops)
		if depth == 0 && f.Pointer() != 0 {
			switch a := f.Elem(); a.Kind() {
			case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
				return FmtValue(a.Interface(), depth+1)
			}
		}
	}

	return fmt.Sprint(value)
}

// FmtValue return vk format string.
func FmtValue(value interface{}, depth int) string {
	if value == nil {
		return ""
	}

	switch f := value.(type) {
	case bool:
		return fmtBool(f)
	case object.Attachment:
		return f.ToAttachment()
	case object.JSONObject:
		return f.ToJSON()
	case reflect.Value:
		return fmtReflectValue(f, depth)
	}

	return fmtReflectValue(reflect.ValueOf(value), depth)
}
