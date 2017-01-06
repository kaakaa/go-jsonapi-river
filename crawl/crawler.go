package crawl

import (
	"fmt"

	"github.com/kaakaa/go-jsonapi-river/store"
)

type Crawler interface {
	Crawl(mongo *store.MongoCollection) error
	GetType() string
}

func GetCrawler(name string) (Crawler, error) {
	// TODO: 各crawlerクラスのGetTypeの値を使用するようにする
	switch name {
	case "jenkins":
		return NewJenkinsCrawler("./conf/jenkins_conf.json")
	case "redmine":
		return NewRedmineCrawler("./conf/redmine_conf.json")
	case "githubIssues":
		return NewGithubIssuesCrawler("./conf/githubIssues_conf.json")
	case "githubOrg":
		return NewGithubOrgCrawler("./conf/githubOrg_conf.json")
	default:
		return nil, fmt.Errorf("No crawler is found: %s", name)
	}
}
