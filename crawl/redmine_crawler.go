package crawl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/kaakaa/go-jsonapi-river/store"
)

type RedmineCrawler struct {
	Host     string
	Query    string
	Auth     string
	AuthInfo RedmineAuthInfo
	Projects []string
}

type RedmineAuthInfo struct {
	User  string
	Token string
}

func NewRedmineCrawler(jenkinsConfFile string) (*RedmineCrawler, error) {
	f, err := ioutil.ReadFile(jenkinsConfFile)
	if err != nil {
		return nil, err
	}

	var crawler *RedmineCrawler
	json.Unmarshal(f, &crawler)
	return crawler, nil
}

func (r *RedmineCrawler) Crawl(mongo *store.MongoCollection) error {
	offset := 0.0
	limit := 100.0
	for _, v := range r.Projects {
		err := r.recordFromAPI(v, offset, limit, mongo)
		return err
	}
	return nil
}

func (r *RedmineCrawler) recordFromAPI(projectName string, offset, limit float64, mongo *store.MongoCollection) error {
	url := fmt.Sprintf("%s/projects/%s/issues.json?offset=%v&limit=%v", r.Host, projectName, offset, limit)
	if len(r.Query) != 0 {
		url += "&" + r.Query
	}

	log.Printf("request to RedmineAPI: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	r.setAuthInfo(req)

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
	issues := m["issues"].([]interface{})
	for _, i := range issues {
		issue := i.(map[string]interface{})
		issue["_id"] = issue["id"]
		mongo.Write(issue)
	}

	totalCount := m["total_count"].(float64)
	offset = m["offset"].(float64)
	limit = m["limit"].(float64)
	leftCount := totalCount - (offset + limit)
	log.Printf("totalCount: %v, offset: %v, limit: %v, leftCount: %v", totalCount, offset, limit, leftCount)
	if leftCount < 0 {
		return nil
	} else {
		time.Sleep(15 * time.Second)
		err = r.recordFromAPI(projectName, offset+limit, limit, mongo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RedmineCrawler) setAuthInfo(req *http.Request) {
	switch r.Auth {
	case "basic":
		req.SetBasicAuth(r.AuthInfo.User, r.AuthInfo.Token)
	}
}

func (c RedmineCrawler) GetType() string {
	return "redmine"
}
