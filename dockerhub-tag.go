package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
)

var Version string = "dev"

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
}

func fatal(err error) {
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}

// Client provides access to manage automated builds on https://hub.docker.com.
type Client struct {
	token string
	repo  string
	cli   *http.Client
}

// NewClient creates a Client to https://hub.docker.com imitating a
// user logging in, with a real browser.
func NewClient(username, password, repository string) *Client {
	c := &Client{
		repo: repository,
		cli:  new(http.Client),
	}

	login := fmt.Sprintf(`{"username": "%s", "password":"%s"}`, username, password)
	resp, err := c.cli.Post("https://hub.docker.com/v2/users/login",
		"application/json", strings.NewReader(login))
	fatal(err)

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var token map[string]string
	fatal(dec.Decode(&token))

	c.token = token["token"]
	log.Debug("JWT:", c.token)

	return c
}

func (c *Client) eachTag(f func(id int, name, sourceType, sourceName, dockerfileLoc string)) {
	c.eachTagAtPage(1, f)
}

func (c *Client) eachTagAtPage(page int, f func(id int, name, sourceType, sourceName, dockerfileLoc string)) {

	resp, err := c.cli.Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/autobuild/tags/?page=%d", c.repo, page))
	fatal(err)
	defer resp.Body.Close()
	list, err := ioutil.ReadAll(resp.Body)
	fatal(err)

	var data struct {
		Count    int
		Next     string
		Previous string
		Results  []struct {
			Id                  int
			Name                string
			Dockerfile_location string
			Source_name         string
			Source_type         string
		}
	}

	err = json.Unmarshal(list, &data)
	fatal(err)

	if data.Next != "" {
		nextUrl, err := url.Parse(data.Next)
		if err != nil {
			log.Warning("Pagination found. Coudnt parse next url:", data.Next)
		}
		nextPage := nextUrl.Query().Get("page")
		p, err := strconv.Atoi(nextPage)
		if err != nil {
			log.Warning("Pagination found. Coudnt get page query param:", data.Next)
		}
		c.eachTagAtPage(p, f)

		log.Debug("Pagination found, next page:", p)

	}
	log.Debug("Number of automated builds:", data.Count)
	for _, tag := range data.Results {
		f(tag.Id, tag.Name, tag.Source_type, tag.Source_name, tag.Dockerfile_location)
	}

}

func (c *Client) ListAll() {
	fmt.Printf("%7s %-30s %-10s %-6s %-10s %-10s\n", "ID", "REPOSITORY", "TAG", "TYPE", "GIT_REF", "DOCKERFILE")
	c.eachTag(func(id int, name, sourceType, sourceName, dockerfileLoc string) {
		fmt.Printf("%7d %-30s %-10s %-6s %-10s %-10s\n", id, c.repo, name, sourceType, sourceName, dockerfileLoc)
	})
}

func (c *Client) deleteAllTag() {

	c.eachTag(func(id int, name, sourceType, sourceName, dockerfileLoc string) {
		if sourceType == "Branch" {
			log.Debug("Ignore branch:", name)
		} else {
			log.Infof("deleting: %8d %-20s", id, name)
			err := c.deleteById(id)
			if err != nil {
				log.Errorf("Unable to delete tag %s: %s", name, err)
			}
		}

	})

}

func (c *Client) AddSingleTag(dockerTag, gitTag, location string) error {
	c.deleteAllTag()
	return c.AddTag(dockerTag, gitTag, location)
}

// AddTag creates a new automated build. The gitTag argument can be any git reference:
// tag or branch. Location is the Dockerfile path inside of the repo. Use "/" if it is
// placed in root. The dockerTag argument is used for image name: myself/myrepo:dockerTag
func (c *Client) AddTag(dockerTag, gitTag, location string) error {

	repo := strings.Split(c.repo, "/")
	data := map[string]string{
		"source_type":         "Tag",
		"source_name":         gitTag,
		"dockerfile_location": location,
		"name":                dockerTag,
		"isNew":               "true",
		"namespace":           repo[0],
		"repoName":            repo[1],
	}

	tagJson, err := json.Marshal(&data)
	fatal(err)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/autobuild/tags/", c.repo),
		bytes.NewBuffer(tagJson),
	)
	fatal(err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("JWT %s", c.token))
	resp, err := c.cli.Do(req)
	if resp.StatusCode != 201 {
		log.Error("Couldn't create tag:", resp.Status)
	}
	fatal(err)

	log.Infof("New automated build created: %s:%s git-tag: %s Dockerfile-loc: %s", c.repo, dockerTag, gitTag, location)
	return nil
}

// DeleteTag deletes an automated build. Please note that it doesn't
// influences existing image tags. So if you delete the v1 automated
// build, than myrepo/myimage:v1 docker image will remain, only it
// will not be built again by Dockerhub.
func (c *Client) DeleteTag(dockerTag string) error {
	c.eachTag(func(id int, name, sourceType, sourceName, dockerfileLoc string) {
		if name == dockerTag {
			log.Infof("Deleting [%d] DockerHub: %s:%s git-ref:%s/%s dockerfile loc: %s\n", id, c.repo, name, sourceType, sourceName, dockerfileLoc)
			c.deleteById(id)
		}
	})

	return nil
}

func (c *Client) deleteById(tagId int) error {

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/autobuild/tags/%d/", c.repo, tagId),
		nil,
	)
	fatal(err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("JWT %s", c.token))
	resp, err := c.cli.Do(req)
	fatal(err)

	log.Debug("Delete resp:", resp.StatusCode)
	if resp.StatusCode != 204 {
		return errors.New(resp.Status)
	}

	return nil
}

func main() {
	usage := `Manage Automated Builds on DockerHub

Usage:
  dockerhub-tag list   <dockerRepo>                                   [--verbose|-v]
  dockerhub-tag add    <dockerRepo> <dockerTag> <gitTag> <location>   [--verbose|-v]
  dockerhub-tag set    <dockerRepo> <dockerTag> <gitTag> <location>   [--verbose|-v]
  dockerhub-tag delete <dockerRepo> <dockerTag>                       [--verbose|-v]

Options:
  -h --help         this message
  -v --verbose      verbose mode`

	ver := fmt.Sprintf("DockerHub Tagger: %s", Version)
	args, _ := docopt.Parse(usage, nil, true, ver, false)

	if args["--verbose"].(bool) {
		log.SetLevel(log.DebugLevel)
	}

	username := os.Getenv("DOCKERHUB_USERNAME")
	password := os.Getenv("DOCKERHUB_PASSWORD")

	if username == "" || password == "" {
		fatal(errors.New("Please set DOCKERHUB_USERNAME and DOCKERHUB_PASSWORD"))
	}
	dhc := NewClient(
		username,
		password,
		args["<dockerRepo>"].(string),
	)

	if args["add"].(bool) {
		err := dhc.AddTag(
			args["<dockerTag>"].(string),
			args["<gitTag>"].(string),
			args["<location>"].(string),
		)
		if err != nil {
			log.Fatal("Cloudn't create tag", err)
		}
	}

	if args["set"].(bool) {
		err := dhc.AddSingleTag(
			args["<dockerTag>"].(string),
			args["<gitTag>"].(string),
			args["<location>"].(string),
		)
		if err != nil {
			log.Fatal("Cloudn't create tag", err)
		}
	}

	if args["list"].(bool) {
		dhc.ListAll()
	}

	if args["delete"].(bool) {
		dhc.DeleteTag(
			args["<dockerTag>"].(string),
		)
	}

}
