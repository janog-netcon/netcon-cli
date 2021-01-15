package vmms

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type client struct {
	Endpoint   string
	Credential string
}

// NewClient vm-management-serverのクライアントを返す
func NewClient(endpoint, credential string) *client {
	return &client{
		Endpoint:   endpoint,
		Credential: credential,
	}
}

type HogeHugaPiyoResponse struct {
}

// HogeHugaPiyo サンプル
func (c *client) HogeHugaPiyo() (*HogeHugaPiyoResponse, error) {
	u := fmt.Sprintf("%s/hogehugapiyo", c.Endpoint)

	cli := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Credential))

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte

	if _, err := resp.Body.Read(body); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hogeHugaPiyoResponse HogeHugaPiyoResponse
	if err := json.Unmarshal(body, &hogeHugaPiyoResponse); err != nil {
		return nil, err
	}

	return &hogeHugaPiyoResponse, nil
}
