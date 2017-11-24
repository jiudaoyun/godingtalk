package godingtalk

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"sync"
)

const (
	VERSION = "0.1"

	BASE_URL = "https://oapi.dingtalk.com/"
	TAOBAO_BASE_URL   = "https://eco.taobao.com/router/rest"
)

//DingTalkClient is the Client to access DingTalk Open API
type DingTalkClient struct {
	CorpID      string
	CorpSecret  string
	AgentID     string
	AccessToken string
	HTTPClient  *http.Client
	Cache       Cache
	*sync.RWMutex

	//社交相关的属性
	SnsAppID string
	SnsAppSecret string
	SnsAccessToken string	
}

//Unmarshallable is
type Unmarshallable interface {
	checkError() error
	getWriter() io.Writer
}

//OAPIResponse is
type OAPIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type TaobaoOAPIResponse struct {
	ErrorResponse struct {
		SubMsg  string `json:"sub_msg"`
		Code    int
		SubCode string `json:"sub_code"`
		Msg     string
	} `json:"error_response"`
}

func (data *OAPIResponse) checkError() (err error) {
	if data.ErrCode != 0 {
		err = fmt.Errorf("%d: %s", data.ErrCode, data.ErrMsg)
	}
	return err
}

func (data *OAPIResponse) getWriter() io.Writer {
	return nil
}

func (data *TaobaoOAPIResponse) checkError() (err error) {
	errData := data.ErrorResponse
	if errData.Code != 0 {
		err = fmt.Errorf("code: %d, msg: %s, sub_code: %s, sub_msg: %s", errData.Code, errData.Msg, errData.SubCode, errData.SubMsg)
	}
	return err
}

func (data *TaobaoOAPIResponse) getWriter() io.Writer {
	return nil
}

//AccessTokenResponse is
type AccessTokenResponse struct {
	OAPIResponse
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	Created     int64
}

//CreatedAt is when the access token is generated
func (e *AccessTokenResponse) CreatedAt() int64 {
	return e.Created
}

//ExpiresIn is how soon the access token is expired
func (e *AccessTokenResponse) ExpiresIn() int {
	return e.Expires
}

//JsAPITicketResponse is
type JsAPITicketResponse struct {
	OAPIResponse
	Ticket  string
	Expires int `json:"expires_in"`
	Created int64
}

//CreatedAt is when the ticket is generated
func (e *JsAPITicketResponse) CreatedAt() int64 {
	return e.Created
}

//ExpiresIn is how soon the ticket is expired
func (e *JsAPITicketResponse) ExpiresIn() int {
	return e.Expires
}

//NewDingTalkClient creates a DingTalkClient instance
func NewDingTalkClient(corpID string, corpSecret string) *DingTalkClient {
	c := new(DingTalkClient)
	c.CorpID = corpID
	c.CorpSecret = corpSecret
	c.HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	c.Cache = NewFileCache(".auth_file")
	c.RWMutex = &sync.RWMutex{}
	return c
}

//RefreshAccessToken is to get a valid access token
func (c *DingTalkClient) RefreshAccessToken() error {
	c.RLock()
	var data AccessTokenResponse
	err := c.Cache.Get(&data)
	if err == nil {
		c.AccessToken = data.AccessToken
		c.RUnlock()
		return nil
	}
	c.RUnlock()

	c.Lock()
	defer c.Unlock()

	params := url.Values{}
	params.Add("corpid", c.CorpID)
	params.Add("corpsecret", c.CorpSecret)
	err = c.httpRPC("gettoken", params, nil, &data)
	if err == nil {
		c.AccessToken = data.AccessToken
		data.Expires = data.Expires | 7200
		data.Created = time.Now().Unix()
		err = c.Cache.Set(&data)
	}
	return err
}

//GetJsAPITicket is to get a valid ticket for JS API
func (c *DingTalkClient) GetJsAPITicket() (ticket string, err error) {
	var data JsAPITicketResponse
	cache := NewFileCache(".jsapi_ticket")
	err = cache.Get(&data)
	if err == nil {
		return data.Ticket, err
	}
	err = c.httpRPC("get_jsapi_ticket", nil, nil, &data)
	if err == nil {
		ticket = data.Ticket
		cache.Set(&data)
	}
	return ticket, err
}

//GetConfig is to return config in json
func (c *DingTalkClient) GetConfig(nonceStr string, timestamp string, url string) (map[string]string, error) {
	ticket, err := c.GetJsAPITicket()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"nonceStr":  nonceStr,
		"agentId":   c.AgentID,
		"timeStamp": timestamp,
		"corpId":    c.CorpID,
		"signature": Sign(ticket, nonceStr, timestamp, url),
	}, nil
}

//Sign is 签名
func Sign(ticket string, nonceStr string, timeStamp string, url string) string {
	s := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s", ticket, nonceStr, timeStamp, url)
	return sha1Sign(s)
}
