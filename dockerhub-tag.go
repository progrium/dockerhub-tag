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

// AddTag creates a new automated build. The gitTag argument can be any git reference:
// tag or branch.  The dockerRepo argument defines the full name of the image: myslef/myrepo.
// Location is the Dockerfile path inside of the repo. Use "/" if it is
// placed in root. The dockerTag argument is used for image name: myself/myrepo:dockerTag
func (c *Client) AddTag(dockerRepo, dockerTag, gitTag, location string) error {

	repoUrl := "https://registry.hub.docker.com/u/" + dockerRepo + "/"
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
	log.Debug("trusted_builds:")
	formLines := 0
	selects.Each(func(i int, s *goquery.Selection) {
		formLines++
		name, _ := s.Attr("name")
		selected := s.Find("option[selected=selected]")
		value, _ := selected.Attr("value")

		prefix := name[0 : len(name)-len("source_type")]
		sourceName := vals.Get(prefix + "source_name")
		hubName := vals.Get(prefix + "name")
		vals.Set(name, value)
		if value == "Tag" {
			log.Debugf("  %s [delete] hubtag:%s  git:%s ", prefix, hubName, sourceName)
			vals.Set(prefix+"DELETE", "on")
		} else {
			log.Debugf("  %s [keep]  hubtag:%s  git:%s ", prefix, hubName, sourceName)
		}
	})
	//newId, _ := c.br.Find("#id_trusted_builds-TOTAL_FORMS").Attr("value")
	data := map[string]string{
		"source_type":         "Tag",
		"source_name":         gitTag,
		"dockerfile_location": location,
		"name":                dockerTag,
	}
	for k, v := range data {
		vals.Set("trusted_builds-"+strconv.Itoa(formLines)+"-"+k, v)
	}
	vals.Set("trusted_builds-TOTAL_FORMS", strconv.Itoa(formLines+1))
	log.Debug("Submitting trusted_build:")
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
	log.Infof("Tag created: %s:%s", dockerRepo, dockerTag)
	return nil
}

// DeleteTag deletes an automated build. Please note that it doesn't
// influences existing image tags. So if you delete the v1 automated
// build, than myrepo/myimage:v1 docker image will remain, only it
// will not be built again by Dockerhub.
func (c *Client) DeleteTag(dockerRepo, dockerTag string) error {
	return nil
}

func main() {
	usage := `Usage:
  dockerhub-tag create <dockerRepo> <dockerTag> <gitTag> <location>   [--verbose|-v]
  dockerhub-tag delete <dockerRepo> <dockerTag>                       [--verbose|-v]

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
			args["<dockerRepo>"].(string),
			args["<dockerTag>"].(string),
			args["<gitTag>"].(string),
			args["<location>"].(string),
		)
		if err != nil {
			log.Fatal("Cloudn't create tag", err)
		}
	}

	if args["delete"].(bool) {
		err = dhc.DeleteTag(
			args["<repo>"].(string),
			args["<tagname>"].(string),
		)
		if err != nil {
			log.Fatal("Cloudn't delete tag", err)
		}

	}

}
