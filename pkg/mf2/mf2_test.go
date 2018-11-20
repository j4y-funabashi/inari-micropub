package mf2_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
)

var defaultAuthor = "https://j4y.co/"
var defaultUuid = "abc-123-456"
var defaultUrl = "https://j4y.co/p/123"

func TestAddsDefaultTypeIfNotDefined(t *testing.T) {
	formData := make(map[string][]string)
	mf := mf2.MfFromForm(formData)

	mf.SetDefaults(defaultAuthor, defaultUuid, defaultUrl)
	t.Logf("%v", mf)

	expected := "h-entry"
	checkSliceContains(t, mf.Type, expected)
}

func TestAddsDefaultPublishedIfNotDefined(t *testing.T) {
	formData := make(map[string][]string)
	mf := mf2.MfFromForm(formData)

	mf.SetDefaults(defaultAuthor, defaultUuid, defaultUrl)
	t.Logf("%v", mf)

	if _, exists := mf.Properties["published"]; exists == false {
		t.Fatalf("published was not added %v", mf)
	}
}

func TestAddsDefaultUid(t *testing.T) {
	formData := make(map[string][]string)
	mf := mf2.MfFromForm(formData)

	mf.SetDefaults(defaultAuthor, defaultUuid, defaultUrl)
	t.Logf("%v", mf)

	if _, exists := mf.Properties["uid"]; exists == false {
		t.Fatalf("uid was not added %v", mf)
	}
}

func TestAddsDefaultAuthor(t *testing.T) {
	formData := make(map[string][]string)
	mf := mf2.MfFromForm(formData)

	mf.SetDefaults(defaultAuthor, defaultUuid, defaultUrl)
	t.Logf("%v", mf)

	if _, exists := mf.Properties["author"]; exists == false {
		t.Fatalf("author was not added %v", mf)
	}
}

func TestUsesPublishedIfDefined(t *testing.T) {
	formData := make(map[string][]string)
	formData["published"] = []string{"2006-01-02T15:04:05-07:00"}
	result := mf2.MfFromForm(formData)
	t.Logf("%v", result)

	checkPropertiesContains(t, result.Properties["published"], "2006-01-02T15:04:05-07:00")
}

func TestAddsTypeFromH(t *testing.T) {
	formData := make(map[string][]string)
	formData["h"] = []string{"chicken", "horse"}
	result := mf2.MfFromForm(formData)
	t.Logf("%v", result)

	checkSliceContains(t, result.Type, "h-chicken")
	checkSliceContains(t, result.Type, "h-horse")
}

func TestAddsAPropery(t *testing.T) {
	formData := make(map[string][]string)
	formData["content"] = []string{"chicken", "horse"}

	result := mf2.MfFromForm(formData)
	t.Logf("%v", result)

	checkPropertiesContains(t, result.Properties["content"], "chicken")
	checkPropertiesContains(t, result.Properties["content"], "horse")
}

func TestDoesNotAddToken(t *testing.T) {
	formData := make(map[string][]string)
	formData["access_token"] = []string{"token"}

	result := mf2.MfFromForm(formData)
	t.Logf("%v", result)

	if _, exists := result.Properties["access_token"]; exists {
		t.Fatalf("access_token was not stripped %v", result)
	}
}

func TestStripsBracketsFromKey(t *testing.T) {
	formData := make(map[string][]string)
	formData["content[]"] = []string{"chicken"}

	result := mf2.MfFromForm(formData)
	t.Logf("%v", result)

	checkPropertiesContains(t, result.Properties["content"], "chicken")
}

func TestFeeds(t *testing.T) {
	p := make(map[string][]interface{})
	p["published"] = append(p["published"], "2017-12-28T07:16:28+00:00")
	mf := mf2.MicroFormat{Type: []string{"h-test"}}
	mf.Properties = p

	expected := []string{"all", "201712"}
	res := mf.Feeds()
	if !reflect.DeepEqual(expected, res) {
		t.Fatalf("expected %#v, got %#v", expected, res)
	}
}

func TestConvertingMicroFormatToViewModel(t *testing.T) {

	p := make(map[string][]interface{})
	p["name"] = append(p["name"], "test-name")
	p["summary"] = append(p["summary"], "test-summary")
	p["content"] = append(p["content"], "test-content")
	p["published"] = append(p["published"], "2018-01-28 10:00:00")
	p["updated"] = append(p["updated"], "test-updated")
	p["author"] = append(p["author"], "test-author")
	p["uid"] = append(p["uid"], "test--uid")
	p["url"] = append(p["url"], "test--url")
	p["rsvp"] = append(p["rsvp"], "test-rsvp")
	p["category"] = append(p["category"], "test-category1")
	p["category"] = append(p["category"], "test-category2")
	p["photo"] = append(p["photo"], "test-photo1")
	p["photo"] = append(p["photo"], "test-photo2")

	p["comment"] = append(p["comment"], "test-comment1")
	p["comment"] = append(p["comment"], "test-comment2")
	p["video"] = append(p["video"], "test-video1")
	p["video"] = append(p["video"], "test-video2")

	p["syndication"] = append(p["syndication"], "test-syndication1")
	p["syndication"] = append(p["syndication"], "test-syndication2")
	p["in-reply-to"] = append(p["in-reply-to"], "test-in-reply-to1")
	p["in-reply-to"] = append(p["in-reply-to"], "test-in-reply-to2")
	p["location"] = append(p["location"], "test-location")
	p["like-of"] = append(p["like-of"], "test-like-of")
	p["repost-of"] = append(p["repost-of"], "test-repost-of")
	p["bookmark-of"] = append(p["bookmark-of"], "test-bookmark-of")
	mf := mf2.MicroFormat{Type: []string{"h-test"}}
	mf.Properties = p

	res := mf.ToView()

	if res.Type != "test" {
		t.Fatalf("jf2 type should be test got '%s'", res.Type)
	}
	if res.Published.Format(time.RFC3339) != "2018-01-28T10:00:00Z" {
		t.Fatalf("jf2 published should be 2018-01-28 10:00:00 +0000 UTC got '%s'", res.Published.Format(time.RFC3339))
	}
	if res.Location != "test-location" {
		t.Fatalf("jf2 location should be test-location got '%s'", res.Location)
	}
	if res.Name != "test-name" {
		t.Fatalf("jf2 name should be test-name got '%s'", res.Name)
	}
	if res.Summary != "test-summary" {
		t.Fatalf("jf2 summary should be test-summary got '%s'", res.Summary)
	}
	if res.Updated != "test-updated" {
		t.Fatalf("jf2 updated should be test-updated got '%s'", res.Updated)
	}
	if res.Author != "test-author" {
		t.Fatalf("jf2 author should be test-author got '%s'", res.Author)
	}
	if res.Content != "test-content" {
		t.Fatalf("jf2 content should be test-content got '%s'", res.Content)
	}
	if res.Rsvp != "test-rsvp" {
		t.Fatalf("jf2 rsvp should be test-rsvp got '%s'", res.Rsvp)
	}
	if res.Uid != "test--uid" {
		t.Fatalf("jf2 uid should be test--uid got '%s'", res.Uid)
	}
	if res.Url != "test--url" {
		t.Fatalf("jf2 url should be test--url got '%s'", res.Url)
	}
	checkSliceContains(t, res.Category, "test-category1")
	checkSliceContains(t, res.Category, "test-category2")
	checkSliceContains(t, res.RepostOf, "test-repost-of")
	checkSliceContains(t, res.LikeOf, "test-like-of")
	checkSliceContains(t, res.BookmarkOf, "test-bookmark-of")

	checkSliceContains(t, res.Syndication, "test-syndication2")
	checkSliceContains(t, res.Syndication, "test-syndication2")
	checkSliceContains(t, res.InReplyTo, "test-in-reply-to2")
	checkSliceContains(t, res.InReplyTo, "test-in-reply-to2")

	checkSliceContains(t, res.Photo, "test-photo1")
	checkSliceContains(t, res.Photo, "test-photo2")

	checkSliceContains(t, res.Comment, "test-comment1")
	checkSliceContains(t, res.Comment, "test-comment2")
	checkSliceContains(t, res.Video, "test-video1")
	checkSliceContains(t, res.Video, "test-video2")
}

func TestRenderHtml(t *testing.T) {
	pubtime, err := time.Parse(time.RFC3339, "2018-01-28T10:00:00Z")
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}

	child1 := mf2.MicroFormatView{
		Type:      "entry",
		Uid:       "test--uid",
		Url:       "/p/test--uid",
		Published: pubtime,
		Author:    "https://j4y.co",
		Photo:     []string{"test--childphoto1", "test--photo2"},
	}
	SUT := mf2.MicroFormatView{
		Type:       "entry",
		Uid:        "test--uid",
		Url:        "/p/test--uid",
		Published:  pubtime,
		Author:     "https://j4y.co",
		LikeOf:     []string{"https://test--likeof"},
		BookmarkOf: []string{"https://test--bookmark"},
		InReplyTo:  []string{"https://test--reply-to"},
		RepostOf:   []string{"https://test--repost-of"},
		Name:       "test--name",
		Content:    "test--content",
		Location:   "test--location",
		Photo:      []string{"test--photo1", "test--photo2"},
		Category:   []string{"test--category1", "test--category2"},
		Children:   []mf2.MicroFormatView{child1},
	}
	var b bytes.Buffer
	err = SUT.Render(&b, "img-proxy")
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}
	expected := ""
	if expected != b.String() {
		t.Fatalf("expected '%s', got '%s'", expected, b.String())
	}
}

func TestSortingChildren(t *testing.T) {
	pubtime, err := time.Parse(time.RFC3339, "2018-01-28T10:00:00Z")
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}

	child1 := mf2.MicroFormatView{
		Type:      "entry",
		Uid:       "test--uid",
		Published: pubtime,
	}
	child2 := mf2.MicroFormatView{
		Type:      "entry",
		Uid:       "test--uid",
		Published: pubtime.Add(1 * time.Hour),
	}
	child3 := mf2.MicroFormatView{
		Type:      "entry",
		Uid:       "test--uid",
		Published: pubtime.Add(2 * time.Hour),
	}
	SUT := mf2.MicroFormatView{
		Type:     "entry",
		Uid:      "test--uid",
		Children: []mf2.MicroFormatView{child3, child1, child2},
	}
	SUT.SortChildren()

	expected := mf2.MicroFormatView{
		Type:     "entry",
		Uid:      "test--uid",
		Children: []mf2.MicroFormatView{child3, child2, child1},
	}
	if !reflect.DeepEqual(expected, SUT) {
		t.Fatalf("expected %+v, got %+v", expected, SUT)
	}
}

func TestCreatingMicroformatFromJSON(t *testing.T) {

	tt := []struct {
		inputJSON   string
		expectedMF2 mf2.MicroFormat
	}{
		{
			inputJSON:   `{"type": ["h-entry"], "properties": {"content": [{"html": "<p>This post has <b>bold</b> and <i>italic</i> text.</p>"}]}}`,
			expectedMF2: mf2.MicroFormat{Type: []string{"h-entry"}, Properties: createProperty("content", map[string]interface{}{"html": "<p>This post has <b>bold</b> and <i>italic</i> text.</p>"})},
		},
		{
			inputJSON:   `{"type": ["h-entry"], "properties": {"content": ["hell"]}, "children": ["https://yoursite.com/photopost1", "https://yoursite.com/photopost2"]}`,
			expectedMF2: mf2.MicroFormat{Type: []string{"h-entry"}, Properties: createProperty("content", "hell"), Children: []interface{}{"https://yoursite.com/photopost1", "https://yoursite.com/photopost2"}},
		},
	}

	for _, tc := range tt {
		SUT, _ := mf2.MfFromJson(tc.inputJSON)
		if !reflect.DeepEqual(tc.expectedMF2, SUT) {
			t.Fatalf("\nexpected %#v,\ngot      %#v", tc.expectedMF2, SUT)
		}
		t.Logf("MF2 ::: %+v", SUT)
	}
}

func createProperty(key string, properties interface{}) map[string][]interface{} {
	p := make(map[string][]interface{})
	p[key] = append(p[key], properties)
	return p
}

func checkPropertiesContains(t *testing.T, properties []interface{}, expected string) {
	for _, v := range properties {
		ty, ok := v.(string)
		if ok && ty == expected {
			return
		}
	}
	t.Fatalf("%s not found in properties %v", expected, properties)
}

func checkSliceContains(t *testing.T, slice []string, expected string) {
	for _, v := range slice {
		if v == expected {
			return
		}
	}
	t.Fatalf("%s not found in slice %v", expected, slice)
}
