package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/browser"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
}

// A Clinet provides access to manage automated builds on https://hub.docker.com.
type Client struct {
	br *browser.Browser
}

// NewClient creates a Client to https://hub.docker.com imitating a
// user logging in, with a real browser.
func NewClient(username, password string) (*Client, error) {
	c := &Client{}
	c.br = surf.NewBrowser()
	err := c.br.Open("https://hub.docker.com/account/login/")
	if err != nil {
		return nil, err
	}
	fm, err := c.br.Form("#form-login")
	if err != nil {
		return nil, err
	} else {
		log.Debug("Login form found")
	}

	fm.Input("username", username)
	fm.Input("password", password)
	err = fm.Submit()

	if err != nil {
		return nil, err
	} else {
		log.Debug("Login success")
	}
	return c, nil
}

// AddTag creates a new automated build. The branch argument can be any git reference:
// tag or branch.  Repo argument defines the full name of the image: myslef/myrepo.
// Location is the Dockerfile path inside of the repo. Use "/" if it is
// placed in root. The tagName argument is used for image name: myself/myrepo:tagName
func (c *Client) AddTag(repo, tagName, branch, location string) error {

	repoUrl := "https://registry.hub.docker.com/u/" + repo + "/"
	err := c.br.Open(repoUrl)
	if err != nil {
		return fmt.Errorf("Couldn't open repo page:%s, error:%s", repoUrl, err)
	} else {
		log.Debug("Repo page opened successfully:", repoUrl)
	}
	err = c.br.Click("ul.nav-tabs li:nth-child(3) a")
	if err != nil {
		return fmt.Errorf("Couldn't navigate to 'Build details' tab! Is it an automated image? error:%s", err)
	}
	//fmt.Println(c.br.Body())
	err = c.br.Click("#repo-info-tab div.repository a:first-child")
	if err != nil {
		return fmt.Errorf("Couldn't navigate to 'Edit Build Details' error:%s", err)
	}

	vals := url.Values{}
	inputs := c.br.Find("#mainform input")
	inputs.Each(func(i int, s *goquery.Selection) {
		type_, _ := s.Attr("type")
		if type_ == "submit" {
			return
		}
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		checked, _ := s.Attr("checked")
		if checked == "checked" {
			value = "on"
		}
		if value == "" {
			return
		}
		vals.Set(name, value)
	})
	selects := c.br.Find("#mainform select")
	selects.Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		selected := s.Find("option[selected=selected]")
		value, _ := selected.Attr("value")
		vals.Set(name, value)
	})
	newId, _ := c.br.Find("#id_trusted_builds-TOTAL_FORMS").Attr("value")
	data := map[string]string{
		"source_type":         "Branch",
		"source_name":         branch,
		"dockerfile_location": location,
		"name":                tagName,
	}
	for k, v := range data {
		vals.Set("trusted_builds-"+newId+"-"+k, v)
	}
	i, err := strconv.ParseInt(newId, 10, 0)
	i = i + 1
	vals.Set("trusted_builds-TOTAL_FORMS", strconv.Itoa(int(i)))
	log.Debug("trusted_build form fields:")
	for k, v := range vals {
		log.Debugf("  %s=%s", k, v)
	}

	log.Debug("Cookies")
	for _, v := range c.br.SiteCookies() {
		log.Debugf("  %s: %s", v.Name, v.Value)
	}
	//time.Sleep(1 * time.Second)
	err = c.br.PostForm(c.br.Url().String(), vals)
	//fm, err = c.br.Form("#mainform")
	if err != nil {
		return fmt.Errorf("Posting new tag failed, error:%s", err)
	}
	//fm.Submit()
	return nil
}

func main() {
	usage := `Usage:
  dockerhub-tag create <repo> <tagname> <branch> <location>   [--verbose|-v]
  dockerhub-tag delete <repo> <tagname>                       [--verbose|-v]

Options:
  -h --help         this message
  -v --verbose      verbose mode`

	ver := fmt.Sprintf("DockerHub Tagger: 0.1.0")
	args, _ := docopt.Parse(usage, nil, true, ver, false)

	if args["--verbose"].(bool) {
		log.SetLevel(log.DebugLevel)
	}

	dhc, err := NewClient(
		os.Getenv("DOCKERHUB_USERNAME"),
		os.Getenv("DOCKERHUB_PASSWORD"),
	)

	if args["create"].(bool) {
		err = dhc.AddTag(
			args["<repo>"].(string),
			args["<tagname>"].(string),
			args["<branch>"].(string),
			args["<location>"].(string),
		)
		if err != nil {
			panic(err)
		}
	}

}
