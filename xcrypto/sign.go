package xcrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xutils"
	"strings"
	"time"
)

type SignClient struct {
	APPID       string
	Debug       bool
	privateKey  *rsa.PrivateKey
	wxPublicKey *rsa.PublicKey
}

func NewSignClient() *SignClient {
	return &SignClient{}
}

const Authorization = "SHA256-RSA2048"

func (c *SignClient) authorization(method, path string, bm xutils.BodyMap) (string, error) {
	var (
		jb        = ""
		timestamp = time.Now().Unix()
		nonceStr  = xutils.RandomString(32)
	)
	if bm != nil {
		jb = bm.JsonBody()
	}
	if strings.HasSuffix(path, "?") {
		path = path[:len(path)-1]
	}
	ts := xutils.Int642String(timestamp)
	_str := method + "\n" + path + "\n" + ts + "\n" + nonceStr + "\n" + jb + "\n"
	if c.Debug {
		logrus.Debugf("signature string:\n%s", _str)
	}
	sign, err := c.rsaSign(_str)
	if err != nil {
		return "", err
	}
	return Authorization + ` appid="` + c.APPID + `",nonce_str="` + nonceStr + `",timestamp="` + ts + `",signature="` + sign + `"`, nil
}

// 私钥签名
func (c *SignClient) rsaSign(str string) (string, error) {
	if c.privateKey == nil {
		return "", errors.New("privateKey can't be nil")
	}
	h := sha256.New()
	h.Write([]byte(str))
	result, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", fmt.Errorf("[%w]: %+v", errors.New("signature error"), err)
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

type signInfo struct {
	HeaderTimestamp string `json:"timestamp"`
	HeaderNonce     string `json:"nonce"`
	HeaderSignature string `json:"signature"`
	SignBody        string `json:"sign_body"`
}

// 公钥验签
func (c *SignClient) verifySyncSign(si *signInfo) error {
	if si == nil {
		return errors.New("auto verify sign, bug SignInfo is nil")
	}
	str := si.HeaderTimestamp + "\n" + si.HeaderNonce + "\n" + si.SignBody + "\n"
	signBytes, _ := base64.StdEncoding.DecodeString(si.HeaderSignature)

	h := sha256.New()
	h.Write([]byte(str))
	if err := rsa.VerifyPKCS1v15(c.wxPublicKey, crypto.SHA256, h.Sum(nil), signBytes); err != nil {
		return fmt.Errorf("[%w]: %v", errors.New("verify signature error"), err)
	}
	return nil
}

// https://pay.weixin.qq.com/wiki/doc/apiv3/wechatpay/wechatpay3_3.shtml
