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
    "strconv"
    "os"
    "encoding/base64"
)

func getBaseURL() string {
    if os.Getenv("BH_ENV") != "" {
        return "http://127.0.0.1:9292"
    }
    return "https://deeq.bithavoc.io"
}


type Client struct {
    client *http.Client
    token id.Token
    ApplicationVersion string
    ApplicationUpgrade UpgradeInfo
    ApplicationUpgradeChanged func(client *Client)
}

type UpgradeInfo struct {
    Available bool
    Version string
    Message string
}

func NewClient(token id.Token) *Client {
    c := &Client {
        client: &http.Client{},
        token: token,
    }
    return c
}

func (client *Client)perform(method string, path string, form url.Values, resultObject interface{}) (resp *http.Response, err error) {
    req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", getBaseURL(), path), strings.NewReader(form.Encode()))
    if err != nil {
        return nil, err
    }
    if method != "GET" {
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    }
    req.Header.Add("X-BITHAVOC-TOKEN", client.token.Code)
    if client.ApplicationVersion != "" {
        req.Header.Add("X-DEEQ-VERSION", "name:CLI,version:" + client.ApplicationVersion)
    }
    resp, err = client.client.Do(req)
    if err != nil {
        return nil, err
    }

    if upgradeString := resp.Header.Get("X-DEEQ-UPGRADE"); upgradeString != "" {
        upgradeInfo := strings.Split(upgradeString, ",")
        latest := strings.Split(upgradeInfo[0], ":")[1]
        message := strings.Split(upgradeInfo[1], ":")[1]
        if data, err := base64.StdEncoding.DecodeString(message); err == nil {
            message = string(data)
        } else {
            message = ""
        }
        client.ApplicationUpgrade = UpgradeInfo {
            Available: true,
            Version: latest,
            Message: message,
        }
        if client.ApplicationUpgradeChanged != nil {
            client.ApplicationUpgradeChanged(client)
        }
        //fmt.Printf("%+v", client.ApplicationUpgrade)
    }

    body := resp.Body
    defer func() {
        if body != nil {
            body.Close()
        }
    }()
    resultData, err := ioutil.ReadAll(resp.Body)
    //fmt.Println(string(resultData))
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
    return TaskId(strings.ToLower(randomString(10)))
}

type baseResult struct {
    Message string
}

const (
    TaskStatusIncomplete int = 0
    TaskStatusComplete= 1
)

type Task struct {
    Id TaskId `json:"task_id"`
    Text string
    Status int
    Deleted bool
}

type taskResult struct {
    baseResult
    Task Task
}

func (task *Task)ToForm() url.Values {
    form :=  url.Values{}
    form.Set("task_id", string(task.Id))
    form.Set("text", task.Text)
    form.Set("status", strconv.Itoa(task.Status))
    if task.Deleted {
        form.Set("deleted", "true")
    } else {
        form.Set("deleted", "false")
    }
    return form
}

type DeeqError struct {
    msg string
}

func (de *DeeqError)Error() string {
    return de.msg
}

func (client *Client) SetTask(task *Task) (*Task, error) {
    form := task.ToForm()
    resultObject := taskResult{}
    _, err := client.perform("POST", "tasks", form, &resultObject)
    if err != nil {
        return nil, err
    }
    if resultObject.Message != "" {
        return nil, &DeeqError{resultObject.Message}
    }
    return &resultObject.Task, nil
}

func (client *Client) GetTask(tid TaskId) (*Task, error) {
    resultObject := taskResult{}
    _, err := client.perform("GET", "tasks/" + string(tid), url.Values{}, &resultObject)
    if err != nil {
        return nil, err
    }
    if resultObject.Message != "" {
        return nil, &DeeqError{resultObject.Message}
    }
    return &resultObject.Task, nil
}


type listResult struct {
    baseResult
    Tasks []Task
}

func (client *Client) GetTasksInTags(rootTag string, childTag string) ([]Task, error) {
    if strings.HasPrefix(rootTag, "#") {
        rootTag = rootTag[1:]
    }
    resultObject := listResult{}
    _, err := client.perform("GET", fmt.Sprintf("tag/%s/tasks", rootTag), url.Values{}, &resultObject)
    if err != nil {
        return nil, err
    }
    if resultObject.Message != "" {
        return nil, &DeeqError{resultObject.Message}
    }
    return resultObject.Tasks, nil
}
