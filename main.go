package main

import (
	"flag"

	"github.com/kaakaa/go-jsonapi-river/crawl"
	"github.com/kaakaa/go-jsonapi-river/store"
)

var dataSource store.DataSource

var (
	jenkinsOpt      = flag.Bool("jenkins", false, "crawl jenkins (needs 'conf/jenkins_conf.json')")
	redmineOpt      = flag.Bool("redmine", false, "crawl redmine (needs 'conf/redmine_conf.json')")
	githubIssuesOpt = flag.Bool("githubIssues", false, "crawl github_issues (needs 'conf/githubIssues_conf.json')")
	githubOrgOpt    = flag.Bool("githubOrg", false, "crawl github_organizations (needs 'conf/githubOrg_conf.json')")
)

func main() {
	flag.Parse()

	var crawlers []crawl.Crawler
	if *jenkinsOpt {
		c, err := crawl.GetCrawler("jenkins")
		if err != nil {
			panic(err)
		}
		crawlers = append(crawlers, c)
	}
	if *redmineOpt {
		c, err := crawl.GetCrawler("redmine")
		if err != nil {
			panic(err)
		}
		crawlers = append(crawlers, c)
	}
	if *githubIssuesOpt {
		c, err := crawl.GetCrawler("githubIssues")
		if err != nil {
			panic(err)
		}
		crawlers = append(crawlers, c)
	}
	if *githubOrgOpt {
		c, err := crawl.GetCrawler("githubOrg")
		if err != nil {
			panic(err)
		}
		crawlers = append(crawlers, c)
	}

	for _, c := range crawlers {
		mongo, err := store.NewMongo("10.25.165.168:27017", "test", c.GetType())
		if err != nil {
			panic(err)
		}
		err = c.Crawl(mongo)
		if err != nil {
			panic(err)
		}
	}
}
