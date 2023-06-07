package nga

import (
	"fmt"

	"github.com/imroc/req/v3"
)

type NgaClient struct {
	*req.Client
	isLogged bool
}

var Client *NgaClient
var BASE_URL string
var UA string
var COOKIE string

func NewNgaClient() *NgaClient {
	c := req.C().
		SetBaseURL(BASE_URL).
		SetCommonHeader("Cookie", COOKIE).
		SetUserAgent(UA).
		OnBeforeRequest(func(c *req.Client, r *req.Request) error {
			if r.RetryAttempt > 0 { // Ignore on retry.
				return nil
			}

			return nil
		}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if !resp.IsSuccessState() {
				return fmt.Errorf("网络请求失败！错误信息: %s\nRaw dump:\n%s", resp.Err.Error(), resp.Dump())
			}
			return nil
		})

	return &NgaClient{
		Client: c,
	}
}
