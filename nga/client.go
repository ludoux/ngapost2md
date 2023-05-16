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
		SetCommonHeader("Cookie",COOKIE).
		SetUserAgent(UA).
		// EnableDump at the request level in request middleware which dump content into
		// memory (not print to stdout), we can record dump content only when unexpected
		// exception occurs, it is helpful to troubleshoot problems in production.
		OnBeforeRequest(func(c *req.Client, r *req.Request) error {
			if r.RetryAttempt > 0 { // Ignore on retry.
				return nil
			}

			return nil
		}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {

			// Corner case: neither an error response nor a success response,
			// dump content to help troubleshoot.
			if !resp.IsSuccess() {
				return fmt.Errorf("bad response, Err:%s\nRaw dump:\n%s", resp.Err.Error(), resp.Dump())
			}
			return nil
		})

	return &NgaClient{
		Client: c,
	}
}
