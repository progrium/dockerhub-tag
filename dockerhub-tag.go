package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	br := surf.NewBrowser()
	err := br.Open("https://hub.docker.com/account/login/")
	assert(err)
	fm, err := br.Form("#form-login")
	assert(err)
	fm.Input("username", os.Getenv("DOCKERHUB_USERNAME"))
	fm.Input("password", os.Getenv("DOCKERHUB_PASSWORD"))
	err = fm.Submit()
	assert(err)

	repo := os.Getenv("DOCKERHUB_REPO")
	err = br.Open("https://registry.hub.docker.com/u/" + repo + "/")
	assert(err)
	err = br.Click("ul.nav-tabs li:nth-child(3) a")
	assert(err)
	//fmt.Println(br.Body())
	err = br.Click("#repo-info-tab div.repository a:first-child")
	assert(err)

	vals := url.Values{}
	inputs := br.Find("#mainform input")
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
	selects := br.Find("#mainform select")
	selects.Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		selected := s.Find("option[selected=selected]")
		value, _ := selected.Attr("value")
		vals.Set(name, value)
	})
	newId, _ := br.Find("#id_trusted_builds-TOTAL_FORMS").Attr("value")
	data := map[string]string{
		"source_type":         "Branch",
		"source_name":         "foobar",
		"dockerfile_location": "/",
		"name":                "foobar",
	}
	for k, v := range data {
		vals.Set("trusted_builds-"+newId+"-"+k, v)
	}
	i, err := strconv.ParseInt(newId, 10, 0)
	i = i + 1
	vals.Set("trusted_builds-TOTAL_FORMS", strconv.Itoa(int(i)))
	for k, v := range vals {
		fmt.Println(k, v)
	}
	for _, v := range br.SiteCookies() {
		fmt.Println(v.Name, v.Value)
	}
	//time.Sleep(1 * time.Second)
	err = br.PostForm(br.Url().String(), vals)
	//fm, err = br.Form("#mainform")
	assert(err)
	//fm.Submit()
	fmt.Println(br.Body())
}
