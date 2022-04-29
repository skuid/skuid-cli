package nlx

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

func GetFastClientForProxyUrl(uri *fasthttp.URI) (client *fasthttp.Client) {
	if uri != nil {
		client = &fasthttp.Client{
			Dial: fasthttpproxy.FasthttpHTTPDialer(uri.String()),
		}
	} else {
		client = nil
	}

	return
}
