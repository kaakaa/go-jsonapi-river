package crawl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kaakaa/go-jsonapi-river/store"
)

type JenkinsCrawler struct {
	Host     string
	Query    string
	Auth     string
	AuthInfo JenkinsAuthInfo
	Jobs     []string
}

type JenkinsAuthInfo struct {
	User  string
	Token string
}

func NewJenkinsCrawler(jenkinsConfFile string) (*JenkinsCrawler, error) {
	f, err := ioutil.ReadFile(jenkinsConfFile)
	if err != nil {
		return nil, err
	}

	var crawler *JenkinsCrawler
	json.Unmarshal(f, &crawler)
	return crawler, nil
}

func (c *JenkinsCrawler) Crawl(mongo *store.MongoCollection) error {
	fmt.Println(c.Jobs)
	for _, v := range c.Jobs {
		fmt.Println(v)
		if err := c.recordFromAPI(v, mongo); err != nil {
			return err
		}
	}
	fmt.Println(c.Jobs)
	return nil
}

func (c *JenkinsCrawler) recordFromAPI(jobName string, mongo *store.MongoCollection) error {
	url := fmt.Sprintf("%s/job/%s/api/json", c.Host, jobName)
	if len(c.Query) != 0 {
		url += "?" + c.Query
	}

	log.Printf("request to JenkinsAPI: %s", url)
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

	m := d.(map[string]interface{})
	builds := m["allBuilds"].([]interface{})
	for _, b := range builds {
		build := b.(map[string]interface{})
		jobID := c.Host + "_" + jobName
		build["_jobID"] = jobID
		build["_id"] = jobID + "_" + build["displayName"].(string)
		mongo.Write(build)
	}
	return nil
}

func (c *JenkinsCrawler) setAuthInfo(req *http.Request) {
	switch c.Auth {
	case "basic":
		req.SetBasicAuth(c.AuthInfo.User, c.AuthInfo.Token)
	}
}

func (c JenkinsCrawler) GetType() string {
	return "jenkins"
}
