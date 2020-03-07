package mf2

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type MediaMetadata struct {
	Uid         string     `json:"uid"`
	URL         string     `json:"url"`
	FileKey     string     `json:"file_key"`
	FileHash    string     `json:"file_hash"`
	MimeType    string     `json:"mime_type"`
	DateTime    *time.Time `json:"date_time"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	IsPublished bool       `json:"is_published"`
}

type MediaList struct {
	Items  []MediaMetadata `json:"items"`
	Paging *ListPaging     `json:"paging,omitempty"`
}

type ListPaging struct {
	After string `json:"after,omitempty"`
}

func (list *MediaList) Add(item MediaMetadata) {
	list.Items = append(list.Items, item)
}

func MfFromForm(formData map[string][]string) MicroFormat {
	newPost := MicroFormat{
		Properties: make(map[string][]interface{})}

	// add type from h
	if val, ok := formData["h"]; ok {
		for _, v := range val {
			newPost.Type = append(newPost.Type, "h-"+v)
		}
	}
	// build properties
	for k, v := range formData {
		k = strings.Trim(k, "[]")
		if k == "access_token" {
			continue
		}
		for _, val := range v {
			newPost.Properties[k] = append(newPost.Properties[k], val)
		}
	}

	return newPost
}

func MfFromJson(body string) (MicroFormat, error) {
	var mf = MicroFormat{}
	err := json.Unmarshal([]byte(body), &mf)
	if err != nil {
		return mf, err
	}
	return mf, nil
}

type PostList struct {
	Items  []MicroFormat `json:"items"`
	Paging *ListPaging   `json:"paging"`
}

func (list *PostList) Add(item MicroFormat) {
	list.Items = append(list.Items, item)
}

func (list PostList) ToJSON() string {
	b, err := json.Marshal(list)
	if err != nil {
		return ""
	}
	buf := bytes.NewBuffer(b)
	return buf.String()
}

func (list *PostList) Sort() {
	sort.Slice(list.Items, func(a, b int) bool {
		return list.Items[a].ToView().Published > list.Items[b].ToView().Published
	})
}

func (list PostList) FindByURL(rawurl string) MicroFormat {
	out := MicroFormat{}
	for _, item := range list.Items {
		u1, _ := url.Parse(item.GetFirstString("url"))
		u2, _ := url.Parse(rawurl)
		if u1.Path == u2.Path {
			return item
		}
	}
	return out
}

func (list PostList) Slice(from, to int) PostList {
	list.Items = list.Items[from:to]
	return list
}

type MicroFormat struct {
	Type       []string                 `json:"type"`
	Properties map[string][]interface{} `json:"properties"`
}

func (mf *MicroFormat) ApplyUpdate(updates map[string][]interface{}) {
	for k, v := range updates {
		mf.Properties[k] = v
	}
}

func (mf MicroFormat) Feeds() []string {
	ym := parseYearMonth(mf.getFirstString("published"))
	return []string{"all", ym}
}

func parseYearMonth(ym string) string {
	t, err := time.Parse(time.RFC3339, ym)
	if err != nil {
		return "000000"
	}
	return t.Format("200601")
}

func (mf *MicroFormat) SetDefaults(defaultAuthor, uuid, url string) {
	// set default type
	if len(mf.Type) == 0 {
		mf.Type = append(mf.Type, "h-entry")
	}
	if mf.Properties == nil {
		mf.Properties = map[string][]interface{}{}
	}
	if len(mf.Properties["published"]) == 0 {
		now := time.Now()
		outLayout := "2006-01-02T15:04:05-07:00"
		mf.Properties["published"] = append(
			mf.Properties["published"],
			now.Format(outLayout))
	}
	if len(mf.Properties["author"]) == 0 {
		mf.Properties["author"] = append(
			mf.Properties["author"],
			defaultAuthor)
	}
	mf.Properties["uid"] = []interface{}{uuid}
	mf.Properties["url"] = []interface{}{url}
}

func (mf MicroFormat) GetFirstString(key string) string {
	return mf.getFirstString(key)
}

func (mf MicroFormat) GetStringSlice(key string) []string {
	return mf.getStringSlice(key)
}

func (mf MicroFormat) ToView() MicroFormatView {
	out := MicroFormatView{}
	out.Type = strings.Trim(mf.Type[0], "h-")
	out.Uid = mf.getFirstString("uid")
	out.Url = mf.getFirstString("url")
	out.Name = mf.getFirstString("name")
	out.Summary = mf.getFirstString("summary")
	out.Content = mf.parseContentValue()
	out.Category = mf.getStringSlice("category")
	out.Photo = mf.getStringSlice("photo")
	out.Video = mf.getStringSlice("video")
	out.Location = mf.parseLocation()
	out.Author = mf.getFirstString("author")
	out.Published = mf.parsePublishedValue()

	out.Updated = mf.getFirstString("updated")
	out.Rsvp = mf.getFirstString("rsvp")
	out.LikeOf = mf.getStringSlice("like-of")
	out.BookmarkOf = mf.getStringSlice("bookmark-of")
	out.RepostOf = mf.getStringSlice("repost-of")
	out.Syndication = mf.getStringSlice("syndication")
	out.InReplyTo = mf.getStringSlice("in-reply-to")
	out.Comment = mf.getStringSlice("comment")

	ym := parseYearMonth(mf.getFirstString("published"))
	out.Archive = ym
	return out
}

// TODO add getProperty and lowercase propertynames

// Appends a property
func (mf *MicroFormat) AddProperty(k string, v interface{}) {
	if mf.Properties == nil {
		mf.Properties = make(map[string][]interface{})
	}
	if mf.Properties[k] == nil {
		mf.Properties[k] = []interface{}{}
	}
	mf.Properties[k] = append(mf.Properties[k], v)
}

func getFirstString(properties []interface{}) string {
	for _, v := range properties {
		p, ok := v.(string)
		if ok {
			return p
		}
	}
	return ""
}

func (mf MicroFormat) getFirstString(key string) string {
	for _, v := range mf.Properties[key] {
		p, ok := v.(string)
		if ok {
			return p
		}
	}
	return ""
}
func (mf MicroFormat) getStringSlice(key string) []string {
	var o []string
	for _, v := range mf.Properties[key] {
		p, ok := v.(string)
		if ok {
			o = append(o, p)
		}
	}
	return o
}
func (mf MicroFormat) parseContentValue() template.HTML {
	s := mf.getFirstString("content")
	if s != "" {
		return template.HTML(s)
	}

	cleaner := bluemonday.UGCPolicy()
	for _, v := range mf.Properties["content"] {
		if p, ok := v.(map[string]interface{}); ok {
			if o, htmlExists := p["html"].(string); htmlExists {
				return template.HTML(cleaner.Sanitize(o))
			}
		}
	}

	return template.HTML("")
}

func (mf MicroFormat) parseLocation() string {
	for _, v := range mf.Properties["location"] {
		location, ok := v.(map[string]interface{})
		if ok {
			locType, ok := location["type"].([]interface{})
			if ok {
				if getFirstString(locType) == "h-adr" || getFirstString(locType) == "h-card" {
					out := []string{}
					props, ok := location["properties"].(map[string]interface{})
					if ok {
						name, ok := props["name"].([]interface{})
						if ok && getFirstString(name) != "" {
							out = append(out, getFirstString(name))
						}
						locality, ok := props["locality"].([]interface{})
						if ok && getFirstString(locality) != "" {
							out = append(out, getFirstString(locality))
						}
						region, ok := props["region"].([]interface{})
						if ok && getFirstString(region) != "" {
							out = append(out, getFirstString(region))
						}
						country, ok := props["country-name"].([]interface{})
						if ok && getFirstString(country) != "" {
							out = append(out, getFirstString(country))
						}
					}
					return strings.Join(out[:], ", ")
				}
			}
		}
	}
	return ""
}

func (mf MicroFormat) parsePublishedValue() string {
	d := normalizeDate(mf.getFirstString("published"))
	return d.Format(time.RFC3339)
}

type MicroFormatView struct {
	Type        string        `json:"type,omitempty"`
	Uid         string        `json:"uid,omitempty"`
	Url         string        `json:"url,omitempty"`
	Published   string        `json:"published,omitempty"`
	Updated     string        `json:"updated,omitempty"`
	Author      string        `json:"author,omitempty"`
	Name        string        `json:"name,omitempty"`
	Summary     string        `json:"summary,omitempty"`
	Content     template.HTML `json:"content,omitempty"`
	Rsvp        string        `json:"rsvp,omitempty"`
	Location    string        `json:"location,omitempty"`
	RepostOf    []string      `json:"repost_of,omitempty"`
	LikeOf      []string      `json:"like_of,omitempty"`
	BookmarkOf  []string      `json:"bookmark_of,omitempty"`
	Category    []string      `json:"category,omitempty"`
	Syndication []string      `json:"syndication,omitempty"`
	InReplyTo   []string      `json:"in_reply_to,omitempty"`
	Photo       []string      `json:"photo,omitempty"`
	Comment     []string      `json:"comment,omitempty"`
	Video       []string      `json:"video,omitempty"`
	Archive     string        `json:"archive,omitempty"`
}

func (jf2 MicroFormatView) PrepForHugo() MicroFormatView {
	jf2.Url = strings.TrimPrefix(jf2.Url, "https://jay.funabashi.co.uk")
	return jf2
}

func (jf2 *MicroFormatView) PrepImageLinks(imgProxy string) {
	for k, v := range jf2.Photo {
		jf2.Photo[k] = imgProxy + v
	}
}

func (jf2 MicroFormatView) Render(w io.Writer, imgProxy string) error {
	t, err := template.ParseFiles("view/post.html")
	if err != nil {
		return err
	}

	jf2.PrepImageLinks(imgProxy)
	// render header
	var header bytes.Buffer
	if len(jf2.LikeOf) > 0 {
		err = t.ExecuteTemplate(&header, "like", jf2)
	}
	if len(jf2.BookmarkOf) > 0 {
		err = t.ExecuteTemplate(&header, "bookmark", jf2)
	}
	if len(jf2.InReplyTo) > 0 {
		err = t.ExecuteTemplate(&header, "replyto", jf2)
	}
	if len(jf2.RepostOf) > 0 {
		err = t.ExecuteTemplate(&header, "repost", jf2)
	}
	if len(jf2.Name) > 0 {
		err = t.ExecuteTemplate(&header, "name", jf2)
	}

	var body bytes.Buffer
	if len(jf2.Photo) > 0 {
		err = t.ExecuteTemplate(&body, "photo", jf2)
	}
	if len(jf2.Content) > 0 {
		err = t.ExecuteTemplate(&body, "content", jf2)
	}

	var meta bytes.Buffer

	err = t.ExecuteTemplate(&meta, "published", struct{ Url, Published string }{Url: jf2.Url, Published: jf2.Published})
	if err != nil {
		log.Printf("failed to render template: %s", err.Error())
	}

	if len(jf2.Author) > 0 {
		err = t.ExecuteTemplate(&meta, "author", jf2)
	}

	if len(jf2.Category) > 0 {
		err = t.ExecuteTemplate(&meta, "category", jf2)
	}
	if len(jf2.Location) > 0 {
		err = t.ExecuteTemplate(&meta, "location", jf2)
	}

	type PostData struct {
		Header template.HTML
		Body   template.HTML
		Meta   template.HTML
		Post   MicroFormatView
	}
	err = t.ExecuteTemplate(
		w,
		"post",
		PostData{Post: jf2, Header: template.HTML(header.String()), Body: template.HTML(body.String()), Meta: template.HTML(meta.String())},
	)
	if err != nil {
		return err
	}
	return nil
}

func normalizeDate(d string) time.Time {
	formats := []string{

		"2006-01-02 15:04:05 -0700 MST",

		// T time sep
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z0700",
		"2006-01-02T15:04:05Z07",
		"2006-01-02T15:04:05",

		// space time sep
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05Z0700",
		"2006-01-02 15:04:05Z07",
		"2006-01-02 15:04:05",

		// no seconds T time sep
		"2006-01-02T15:04Z07:00",
		"2006-01-02T15:04Z0700",
		"2006-01-02T15:04Z07",

		// no seconds space time sep
		"2006-01-02 15:04Z07:00",
		"2006-01-02 15:04Z0700",
		"2006-01-02 15:04Z07",

		"2006-01-02",
		"2006-01",
		"2006",
	}

	for _, format := range formats {
		t, err := time.Parse(format, d)
		if err == nil {
			return t
		}
	}

	log.Printf("[E] Could not parse date format [ %v ]", d)
	return time.Now()
}
