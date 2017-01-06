package crawl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kaakaa/go-jsonapi-river/store"
)

type GithubIssuesCrawler struct {
	Host     string
	Query    string
	Auth     string
	AuthInfo GithubIssuesAuthInfo
	Repos    []string
}

type GithubIssuesAuthInfo struct {
	User     string
	Password string
}

func NewGithubIssuesCrawler(githubIssuesConfFile string) (*GithubIssuesCrawler, error) {
	f, err := ioutil.ReadFile(githubIssuesConfFile)
	if err != nil {
		return nil, err
	}

	var crawler *GithubIssuesCrawler
	json.Unmarshal(f, &crawler)
	fmt.Println("githubIssues")
	return crawler, nil
}

func (c *GithubIssuesCrawler) Crawl(mongo *store.MongoCollection) error {
	fmt.Println(c.Repos)
	for _, v := range c.Repos {
		fmt.Println(v)
		if err := c.recordFromAPI(v, c.Query, mongo); err != nil {
			return err
		}
	}
	return nil
}

func (c *GithubIssuesCrawler) recordFromAPI(repoName, query string, mongo *store.MongoCollection) error {
	url := fmt.Sprintf("%s/repos/%s/issues", c.Host, repoName)
	if len(query) != 0 {
		url += "?" + query
	}

	log.Printf("request to GithubIssuesAPI: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	c.setAuthInfo(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var d interface{}
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return err
	}

	issues := d.([]interface{})
	for _, i := range issues {
		issue := i.(map[string]interface{})
		issueNumber := fmt.Sprintf("%s_%v", repoName, issue["number"])
		issue["_id"] = issueNumber
		mongo.Write(issue)
	}

	links := resp.Header.Get("Link")
	link := ""
	for _, v := range strings.Split(links, ",") {
		if strings.HasSuffix(v, `rel="next"`) {
			l := strings.Split(v, ";")[0]
			link = l[1 : len(l)-1]
		}
	}
	if len(link) != 0 {
		time.Sleep(15 * time.Second)
		q := strings.Split(link, "?")[1]
		fmt.Println(q)
		c.recordFromAPI(repoName, q, mongo)
	}

	return nil
}

func (c *GithubIssuesCrawler) setAuthInfo(req *http.Request) {
	switch c.Auth {
	case "basic":
		req.SetBasicAuth(c.AuthInfo.User, c.AuthInfo.Password)
	}
}

func (c GithubIssuesCrawler) GetType() string {
	return "githubIssues"
}
