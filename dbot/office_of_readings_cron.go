package dbot

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/christiansoetanto/tbd-bot/domain"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/util"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (u *usecase) officeOfReadingsCronJob(ctx context.Context) func() {
	return func() {
		embed, err := util.GenerateOfficeOfReadingsEmbeds()
		if err != nil {
			return
		}
		for _, config := range u.Config.GuildConfig {
			if config.RegisteredFeature[domain.FeatureKeyOfficeOfReadingsCron] {
				_, err = u.Session.ChannelMessageSendEmbeds(config.Channel.OfficeOfReadings, embed)
				if err != nil {
					logv2.Error(ctx, err)
				}
			}
		}
	}
}

func (u *usecase) officeOfReadingsCronJob2(ctx context.Context) func() {
	return func() {
		ctx = logv2.InitRequestContext(ctx)
		ctx = logv2.InitFuncContext(ctx)
		logv2.Debug(ctx, logv2.Info, "Starting office of readings cron job")
		title, text, err := util.GetOfficeOfReadingsText()
		if err != nil {
			logv2.Error(ctx, err)
			return
		}
		logv2.Debug(ctx, logv2.Info, title, "title")
		logv2.Debug(ctx, logv2.Info, text, "text")
		doSomeGitStuffs(ctx, title, text)
		logv2.Debug(ctx, logv2.Info, "Finished office of readings cron job")
	}
}

func doSomeGitStuffs(ctx context.Context, title, text string) {
	ctx = logv2.InitFuncContext(ctx)

	//branch := time.Now().Format(time.RFC3339)
	branch := time.Now().Format("2006-01-02")
	date := time.Now().Format("2006-01-02")
	if err := retry.Do(func() error {
		return createRef(ctx, branch)
	}); err != nil {
		logv2.Error(ctx, err)
		return
	}

	filename := fmt.Sprintf("%s-%s.md", date, title)
	gitTitle := fmt.Sprintf("add office of readings for %s", date)

	if err := retry.Do(func() error {
		return createFile(ctx, filename, text, branch, gitTitle)
	}); err != nil {
		logv2.Error(ctx, err)

		return
	}
	data, err := retry.DoWithData(func() ([]byte, error) {
		pr, err := createPR(ctx, gitTitle, title, branch)
		if err != nil {
			return nil, err
		}
		return []byte(fmt.Sprintf("%d", pr)), nil
	})
	if err != nil {
		logv2.Error(ctx, err)

		return
	}
	pr, _ := strconv.ParseInt(string(data), 10, 64)
	if err := retry.Do(func() error {
		return mergePR(ctx, int(pr), gitTitle)
	}); err != nil {
		logv2.Error(ctx, err)
		return
	}

	if err := retry.Do(func() error {
		return deleteRef(ctx, branch)
	}); err != nil {
		logv2.Error(ctx, err)
		return
	}
}

const (
	POST   = "POST"
	GET    = "GET"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type apiErrorResponse struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
}

func hit(ctx context.Context, method string, path string, payload any) (string, error) {
	ctx = logv2.InitFuncContext(ctx)
	url := fmt.Sprintf("https://api.github.com/repos/christiansoetanto/office-of-readings/%s", path)
	var p string
	if payload != nil {
		switch payload.(type) {
		case string:
			p = payload.(string)
		default:
			pB, _ := json.Marshal(payload)
			p = string(pB)
		}
	}

	logv2.Debug(ctx, logv2.Request, p, "Payload")
	logv2.Debug(ctx, logv2.Info, url, "URL")

	payloadReader := strings.NewReader(p)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payloadReader)

	if err != nil {
		logv2.Error(ctx, err)
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUBAPITOKEN")))

	res, err := client.Do(req)
	if err != nil {
		logv2.Error(ctx, err)
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logv2.Error(ctx, err)
		return "", err
	}
	logv2.Debug(ctx, logv2.Response, body, "Response")

	if res.StatusCode >= 400 {
		var apiErr apiErrorResponse
		err = json.Unmarshal(body, &apiErr)
		if err != nil {
			return "", err
		}
		whitelistedErrorMessage := []string{
			"Reference already exists",
			"Invalid request.\n\n\"sha\" wasn't supplied.",
		}

		for _, msg := range whitelistedErrorMessage {
			if msg == apiErr.Message {
				return string(body), nil
			}
		}
		return "", errors.New(apiErr.Message)
	}

	return string(body), nil
}

type ref struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

func createRef(ctx context.Context, branchName string) error {
	ctx = logv2.InitFuncContext(ctx)
	path := "git/refs"
	method := POST
	sha, err := getLatestRefSHA(ctx)
	if err != nil {
		return err
	}
	r := ref{
		Ref: fmt.Sprintf("refs/heads/%s", branchName),
		Sha: sha,
	}
	_, err = hit(ctx, method, path, r)
	if err != nil {
		return err
	}
	return nil
}
func getLatestRefSHA(ctx context.Context) (string, error) {
	ctx = logv2.InitFuncContext(ctx)

	path := "git/matching-refs/heads/master"
	method := GET
	resp, err := hit(ctx, method, path, nil)
	if err != nil {
		logv2.Error(ctx, err)
		return "", err
	}
	sha := gjson.Get(resp, "0.object.sha")

	return sha.String(), nil
}

type content struct {
	Content string `json:"content"`
	Message string `json:"message"`
	Branch  string `json:"branch"`
}

func createFile(ctx context.Context, filename string, cont string, branch string, message string) error {
	ctx = logv2.InitFuncContext(ctx)

	path := "contents/" + filename
	method := PUT
	//convert string to base64
	b64 := base64.StdEncoding.EncodeToString([]byte(cont))
	r := content{
		Content: b64,
		Message: message,
		Branch:  branch,
	}
	_, err := hit(ctx, method, path, r)
	if err != nil {
		return err
	}
	return nil
}

type createPull struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

func createPR(ctx context.Context, title, body, branch string) (int, error) {
	ctx = logv2.InitFuncContext(ctx)

	path := "pulls"
	method := POST
	r := createPull{
		Title: title,
		Body:  body,
		Head:  branch,
		Base:  "master",
	}
	s, err := hit(ctx, method, path, r)
	if err != nil {
		return 0, err
	}
	get := gjson.Get(s, "number")
	return int(get.Int()), nil
}

type mergePull struct {
	CommitTitle string `json:"commit_title,omitempty"`
	MergeMethod string `json:"merge_method,omitempty"`
}

func mergePR(ctx context.Context, prNumber int, title string) error {
	ctx = logv2.InitFuncContext(ctx)

	path := fmt.Sprintf("pulls/%d/merge", prNumber)
	method := PUT
	r := mergePull{
		CommitTitle: title,
		MergeMethod: "squash",
	}
	_, err := hit(ctx, method, path, r)
	if err != nil {
		return err
	}
	return nil
}

func deleteRef(ctx context.Context, branchName string) error {
	ctx = logv2.InitFuncContext(ctx)
	path := fmt.Sprintf("git/refs/heads/%s", branchName)
	method := DELETE
	_, err := hit(ctx, method, path, nil)
	if err != nil {
		return err
	}
	return nil
}
