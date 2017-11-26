package godingtalk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
	"fmt"
)

const typeJSON = "application/json"
const typeFormURLEncoded = "application/x-www-form-urlencoded;charset=utf-8"

//UploadFile is for uploading a single file to DingTalk
type UploadFile struct {
	FieldName string
	FileName  string
	Reader    io.Reader
}

//DownloadFile is for downloading a single file from DingTalk
type DownloadFile struct {
	MediaID  string
	FileName string
	Reader   io.Reader
}

func (c *DingTalkClient) httpRPC(path string, params url.Values, requestData interface{}, responseData Unmarshallable) error {
	if params == nil {
		params = url.Values{}
	}
	if c.AccessToken != "" && params.Get("access_token") == "" {
		params.Set("access_token", c.AccessToken)}
	return c.httpRequest(path, params, requestData, responseData)
}

func (c *DingTalkClient) httpTaobaoRPC(method string, params url.Values, responseData Unmarshallable) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("format", "json")
	params.Set("method", method)
	params.Set("partner_id", "apidoc")
	params.Set("session", c.AccessToken)
	params.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	params.Set("v", "2.0")
	params.Set("simplify", "true")

	return c.httpRequest("", params, nil, responseData, true)
}

func (c *DingTalkClient) httpRequest(path string, params url.Values, requestData interface{}, responseData Unmarshallable, throughTaobao... bool) error {
	client := c.HTTPClient
	var request *http.Request

	if len(throughTaobao) == 0 || !throughTaobao[0] {
		url := BASE_URL + path + "?" + params.Encode()
		if requestData != nil {
			switch requestData.(type) {
			case UploadFile:
				var b bytes.Buffer
				request, _ = http.NewRequest("POST", url, &b)
				w := multipart.NewWriter(&b)

				uploadFile := requestData.(UploadFile)
				if uploadFile.Reader == nil {
					return errors.New("upload file is empty")
				}
				fw, err := w.CreateFormFile(uploadFile.FieldName, uploadFile.FileName)
				if err != nil {
					return err
				}
				if _, err = io.Copy(fw, uploadFile.Reader); err != nil {
					return err
				}
				if err = w.Close(); err != nil {
					return err
				}
				request.Header.Set("Content-Type", w.FormDataContentType())
			default:
				d, _ := json.Marshal(requestData)
				// log.Printf("url: %s request: %s", url, string(d))
				request, _ = http.NewRequest("POST", url, bytes.NewReader(d))
				request.Header.Set("Content-Type", typeJSON)
			}
		} else {
			// log.Printf("url: %s", url)
			request, _ = http.NewRequest("GET", url, nil)
		}
	} else {
		buf := strings.NewReader(params.Encode())
		request, _ = http.NewRequest("POST", TAOBAO_BASE_URL, buf)
		request.Header.Set("Content-Type", typeFormURLEncoded)
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Server error: " + resp.Status)
	}

	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("response content type: %s", contentType)
	pos := len(typeJSON)
	if len(contentType) >= pos && contentType[0:pos] == typeJSON {
		content, err := ioutil.ReadAll(resp.Body)
		fmt.Printf("response: %s\n", content)
		if err == nil {
			json.Unmarshal(content, responseData)
			return responseData.checkError()
		}
	} else {
		io.Copy(responseData.getWriter(), resp.Body)
		return responseData.checkError()
	}
	return err
}
