package deeq

import (
    id "github.com/bithavoc/id-go-client"
    "net/http"
    "net/url"
    "fmt"
    "io/ioutil"
    "strings"
    "encoding/json"
    "math/rand"
)

const (
    baseURL = "http://127.0.0.1:9292"
)

type Client struct {
    client *http.Client
    token id.Token
}

func NewClient(token id.Token) *Client {
    c := &Client {
        client: &http.Client{},
        token: token,
    }
    return c
}

func (client *Client)perform(path string, form url.Values, resultObject interface{}) (resp *http.Response, err error) {
    req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", baseURL, path), strings.NewReader(form.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Add("X-BITHAVOC-TOKEN", client.token.Code)
    resp, err = client.client.Do(req)
    if err != nil {
        return nil, err
    }
    body := resp.Body
    defer func() {
        if body != nil {
            body.Close()
        }
    }()
    resultData, err := ioutil.ReadAll(resp.Body)
    err = json.Unmarshal(resultData, &resultObject)
    if err != nil {
        return nil, err
    }
    return resp, err
}

type TaskId string

func randomString(l int ) string {
    bytes := make([]byte, l)
    for i:=0 ; i<l ; i++ {
        bytes[i] = byte(randInt(65,90))
    }
    return string(bytes)
}

func randInt(min int, max int) int {
    return min + rand.Intn(max-min)
}
func NewTaskId() TaskId {
    return TaskId(randomString(10))
}

type setTaskResult struct {
    Task_Id string
    Text string
}

func (client *Client) SetTask(tid TaskId, text string) (error) {
    form :=  url.Values{}
    form.Set("task_id", string(tid))
    form.Set("text", text)

    resultObject := setTaskResult{}
    _, err := client.perform("tasks", form, &resultObject)
    if err != nil {
        return err
    }
    return nil
}

