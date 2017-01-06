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

type GithubOrgCrawler struct {
	Host     string
	Query    string
	Auth     string
	AuthInfo GithubOrgAuthInfo
	Orgs     []string
}

type GithubOrgAuthInfo struct {
	User     string
	Password string
}

func NewGithubOrgCrawler(githubOrgConfFile string) (*GithubOrgCrawler, error) {
	f, err := ioutil.ReadFile(githubOrgConfFile)
	if err != nil {
		return nil, err
	}

	var crawler *GithubOrgCrawler
	json.Unmarshal(f, &crawler)
	fmt.Println("githubOrg")
	return crawler, nil
}

func (c *GithubOrgCrawler) Crawl(mongo *store.MongoCollection) error {
	fmt.Println(c.Orgs)
	for _, v := range c.Orgs {
		fmt.Println(v)
		if err := c.recordFromAPI(v, c.Query, mongo); err != nil {
			return err
		}
	}
	return nil
}

func (c *GithubOrgCrawler) recordFromAPI(orgName, query string, mongo *store.MongoCollection) error {
	url := fmt.Sprintf("%s/orgs/%s/repos", c.Host, orgName)
	if len(query) != 0 {
		url += "?" + query
	}

	log.Printf("request to GithubOrgAPI: %s", url)
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

	repos := d.([]interface{})
	for _, i := range repos {
		repo := i.(map[string]interface{})
		id := fmt.Sprintf("%s/%v", orgName, repo["name"])
		repo["_id"] = id
		mongo.Write(repo)
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
		c.recordFromAPI(orgName, q, mongo)
	}

	return nil
}

func (c *GithubOrgCrawler) setAuthInfo(req *http.Request) {
	switch c.Auth {
	case "basic":
		req.SetBasicAuth(c.AuthInfo.User, c.AuthInfo.Password)
	}
}

func (c GithubOrgCrawler) GetType() string {
	return "githubOrg"
}
