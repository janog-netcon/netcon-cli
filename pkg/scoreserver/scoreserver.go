package scoreserver

/*
https://github.com/janog-netcon/netcon-score-server/blob/janog47-changes/vmdb-api/main.go#L85

e.GET("/problem-environments", listProblemEnvironment)
e.GET("/problem-environments/:name", getProblemEnvironment)
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/janog-netcon/netcon-cli/pkg/types"
	"golang.org/x/xerrors"
)

type client struct {
	Endpoint string
}

// NewClient スコアサーバのクライアントを返す
func NewClient(endpoint string) *client {
	return &client{
		Endpoint: endpoint,
	}
}

// ListProblemEnvironment VM一覧を取得する
func (c *client) ListProblemEnvironment() (*[]types.ProblemEnvironment, error) {
	u := fmt.Sprintf("%s/problem-environments", c.Endpoint)

	cli := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.New("status code not 200")
	}

	var problemEnvironments []types.ProblemEnvironment
	if err := json.Unmarshal(respBody, &problemEnvironments); err != nil {
		return nil, err
	}

	return &problemEnvironments, nil
}

// GetProblemEnvironment nameで指定したVM情報を取得する
// APIからは types.ProblemEnvironment が2つArrayで返ってくる
// (VMには SSH と HTTP の2種類のserviceがあり、それぞれ別扱いになっているため)
// それぞれで値が変わるのは、 `service`, `port`, `host` の3つで、それ以外は同じ値が入っている
//
// [
//   {
//     "id": "21864669-7eed-42df-98e3-e96e2c5857b0",
//     "inner_status": null,
//     "status": "STOPPING",
//     "host": "35.187.220.33",		<-
//     "user": "j47-user",
//     "password": "xxxxx",
//     "problem_id": "4b71d7be-6a76-4a10-a16b-9f50b47c3407",
//     "created_at": "2021-01-07T21:43:07.13899Z",
//     "updated_at": "2021-01-07T22:06:06.069066Z",
//     "name": "image-110-okaxv",
//     "service": "SSH",			<-
//     "port": 50080,       		<-
//     "machine_image_name": "image-110"
//   },
//   {
//     "id": "71ff4819-fd3d-4ca6-8598-37c2c365c70f",
//     "inner_status": null,
//     "status": "STOPPING",
//     "host": "35.187.220.33",		<- //実際にはFQDNが入っている
//     "user": "j47-user",
//     "password": "xxxxx",
//     "problem_id": "4b71d7be-6a76-4a10-a16b-9f50b47c3407",
//     "created_at": "2021-01-07T21:43:07.18566Z",
//     "updated_at": "2021-01-07T22:06:06.09428Z",
//     "name": "image-110-okaxv",
//     "service": "HTTPS",			<-
//     "port": 443,         		<-
//     "machine_image_name": "image-110"
//   }
// ]
func (c *client) GetProblemEnvironment(name string) (*[]types.ProblemEnvironment, error) {
	u := fmt.Sprintf("%s/problem-environments/%s", c.Endpoint, name)

	cli := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.New("status code not 200")
	}

	var problemEnvironments []types.ProblemEnvironment
	if err := json.Unmarshal(respBody, &problemEnvironments); err != nil {
		return nil, err
	}

	return &problemEnvironments, nil
}
