package binding

import (
	"github.com/smallfish-root/common-pkg/xencoding"
	"github.com/smallfish-root/common-pkg/xencoding/form"
	"net/http"
	"net/url"
)

func BindQuery(values url.Values, target interface{}) error {
	return xencoding.GetCodec(form.Name).Unmarshal([]byte(values.Encode()), target)
}

func BindForm(req *http.Request, target interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	return xencoding.GetCodec(form.Name).Unmarshal([]byte(req.Form.Encode()), target)
}
