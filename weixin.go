package weixin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client struct {
	logger *zap.Logger
}

func NewClient(logger *zap.Logger) *Client {
	return &Client{
		logger: logger,
	}
}

func (c *Client) Request(method, url string, query url.Values, body map[string]interface{}) ([]byte, error) {
	var (
		reqBody []byte
	)
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	c.logger.Info("Request", zap.String("URL", req.URL.String()))

	// 发出请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	// 解析响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Request", zap.String("body", string(respBody)))
		err := errors.New(string(respBody))
		return nil, errors.WithStack(err)
	}

	return respBody, nil
}

// corpId 企业ID
// corpSecret 应用Secret
func (c *Client) GetAccessToken(corpId, corpSecret string) (string, error) {
	// 调用接口返回登录信息access_token
	method := http.MethodGet
	path := "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	query := url.Values{}
	query.Add("corpid", corpId)
	query.Add("corpsecret", corpSecret)

	resp, err := c.Request(method, path, query, nil)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return "", errors.WithStack(err)
	}

	var reply struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := errors.New(string(resp))
		return "", errors.WithStack(err)
	}

	return reply.AccessToken, nil
}

func (c *Client) SendMessage(message, agentId, toParty, toTag, accessToken string) error {
	// 发送文本消息
	// 调用接口返回登录信息access_token
	method := http.MethodPost
	path := "https://qyapi.weixin.qq.com/cgi-bin/message/send"
	query := url.Values{}
	query.Add("access_token", accessToken)
	body := map[string]interface{}{
		"msgtype": "text",
		"agentid": agentId,
		"text": map[string]string{
			"content": message,
		},
	}
	if toParty != "" {
		body["toparty"] = toParty
	}
	if toTag != "" {
		body["totag"] = toTag
	}

	resp, err := c.Request(method, path, query, body)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return errors.WithStack(err)
	}

	var reply struct {
		ErrMsg string `json:"errmsg"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := errors.New(string(resp))
		return errors.WithStack(err)
	}

	if reply.ErrMsg != "ok" {
		err := errors.New(string(resp))
		return errors.WithStack(err)
	}
	return nil
}

func (c *Client) TagList(accessToken string) error {
	// 发送文本消息
	// 调用接口返回登录信息access_token
	method := http.MethodGet
	path := "https://qyapi.weixin.qq.com/cgi-bin/tag/list"

	query := url.Values{}
	query.Add("access_token", accessToken)

	resp, err := c.Request(method, path, query, nil)
	if err != nil {
		c.logger.Error(path, zap.Error(err))
		return errors.WithStack(err)
	}

	var reply struct {
		ErrMsg string `json:"errmsg"`
	}

	if err := json.Unmarshal(resp, &reply); err != nil {
		c.logger.Error(path, zap.String("resp", string(resp)), zap.Error(err))
		err := errors.New(string(resp))
		return errors.WithStack(err)
	}

	if reply.ErrMsg != "ok" {
		err := errors.New(string(resp))
		return errors.WithStack(err)
	}
	return nil
}
