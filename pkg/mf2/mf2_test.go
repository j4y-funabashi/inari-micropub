package mf2_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/matryer/is"
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

func TestApplyUpdate(t *testing.T) {
	var tests = []struct {
		name     string
		updates  map[string][]interface{}
		mf       mf2.MicroFormat
		expected mf2.MicroFormat
	}{
		{
			name: "it updates simple lists of properties",
			updates: map[string][]interface{}{
				"content": []interface{}{
					"cheese",
					"horse",
				},
			},
			mf: mf2.MicroFormat{
				Properties: map[string][]interface{}{
					"content": []interface{}{
						"hellchicken",
					},
					"uid": []interface{}{
						"123",
					},
				},
			},
			expected: mf2.MicroFormat{
				Properties: map[string][]interface{}{
					"content": []interface{}{
						"cheese",
						"horse",
					},
					"uid": []interface{}{
						"123",
					},
				},
			},
		},
		{
			name: "it updates nested structures",
			updates: map[string][]interface{}{
				"location": []interface{}{
					map[string]interface{}{
						"type": []interface{}{"h-card"},
						"properties": map[string]interface{}{
							"city": []interface{}{
								"leeds",
							},
						},
					},
				},
			},
			mf: mf2.MicroFormat{
				Properties: map[string][]interface{}{
					"content": []interface{}{
						"hellchicken",
					},
					"uid": []interface{}{
						"123",
					},
				},
			},
			expected: mf2.MicroFormat{
				Properties: map[string][]interface{}{
					"content": []interface{}{
						"hellchicken",
					},
					"uid": []interface{}{
						"123",
					},
					"location": []interface{}{
						map[string]interface{}{
							"type": []interface{}{"h-card"},
							"properties": map[string]interface{}{
								"city": []interface{}{
									"leeds",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mf.ApplyUpdate(tt.updates)
			is.Equal(tt.mf, tt.expected)
		})
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

	mfLoc := make(map[string]interface{})
	mfLoc["type"] = []interface{}{"h-adr"}
	pLoc := make(map[string]interface{})
	pLoc["locality"] = []interface{}{"Leeds"}
	pLoc["region"] = []interface{}{"West Yorkshire"}
	pLoc["country-name"] = []interface{}{"UK"}
	// pLoc["latitude"] = []interface{}{"6.66"}
	// pLoc["longitude"] = append(pLoc["longitude"], "7.77")
	mfLoc["properties"] = pLoc
	p["location"] = append(p["location"], mfLoc)

	p["like-of"] = append(p["like-of"], "test-like-of")
	p["repost-of"] = append(p["repost-of"], "test-repost-of")
	p["bookmark-of"] = append(p["bookmark-of"], "test-bookmark-of")
	mf := mf2.MicroFormat{Type: []string{"h-test"}}
	mf.Properties = p

	res := mf.ToView()

	if res.Type != "test" {
		t.Fatalf("jf2 type should be test got '%s'", res.Type)
	}
	if res.Published != "2018-01-28T10:00:00Z" {
		t.Fatalf("jf2 published should be 2018-01-28 10:00:00 +0000 UTC got '%s'", res.Published)
	}
	if res.Location != "Leeds, West Yorkshire, UK" {
		t.Fatalf("jf2 location should be 'Leeds, West Yorkshire, UK' got '%s'", res.Location)
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
	t.SkipNow()
	pubtime := "2018-01-28T10:00:00Z"

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
	}
	var b bytes.Buffer
	err := SUT.Render(&b, "img-proxy")
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}
	expected := ""
	if expected != b.String() {
		t.Fatalf("expected '%s', got '%s'", expected, b.String())
	}
}

func TestMf2FromJSONToView(t *testing.T) {
	tt := []struct {
		name      string
		inputJSON string
		expected  mf2.MicroFormatView
	}{
		{
			name: "minimal valid h-entry",
			inputJSON: `{
			"type": ["h-entry"],
			"properties": {
				"published": ["2018-01-28T00:00:00Z"]
				}
			}`,
			expected: mf2.MicroFormatView{
				Type:      "entry",
				Published: "2018-01-28T00:00:00Z",
				Archive:   "201801",
			},
		},
		{
			name: "full valid h-entry",
			inputJSON: `{
			"type": ["h-entry"],
			"properties": {
				"uid": ["9a9ecd17-2fcf-4d91-97e2-09e2cd9e06b5"],
				"url": ["https://example.com/test1"],
				"name": ["test-name"],
				"summary": ["test-summary"],
				"content": ["test-content"],
				"category": ["tag1", "tag2"],
				"photo": ["https://media.example.com/test1.jpg", "https://media.example.com/test2.jpg"],
				"video": ["https://media.example.com/test1.mp4", "https://media.example.com/test2.mp4"],
				"location": [
					{
						"properties": {
							"country-name": ["United Kingdom"],
							"latitude": [53.800755],
							"locality": ["Leeds"],
							"longitude": [-1.549077],
							"region": ["West Yorkshire"]
						},
						"type": ["h-adr"]
					}
				],
				"author": ["https://jay.funabashi.co.uk/"],
				"published": ["2018-01-28T00:00:00Z"],
				"updated": ["2018-01-29T00:00:00Z"],
				"rsvp": ["https://example.com/rsvp1"],
				"like-of": ["https://example.com/like1"],
				"bookmark-of": ["https://example.com/bookmark1"],
				"repost-of": ["https://example.com/repost1"],
				"syndication": ["https://example.com/syndicate1"],
				"in-reply-to": ["https://example.com/reply1"],
				"comment": ["https://example.com/comment1"]
				}
			}`,
			expected: mf2.MicroFormatView{
				Type:        "entry",
				Uid:         "9a9ecd17-2fcf-4d91-97e2-09e2cd9e06b5",
				Url:         "https://example.com/test1",
				Name:        "test-name",
				Summary:     "test-summary",
				Content:     "test-content",
				Category:    []string{"tag1", "tag2"},
				Photo:       []string{"https://media.example.com/test1.jpg", "https://media.example.com/test2.jpg"},
				Video:       []string{"https://media.example.com/test1.mp4", "https://media.example.com/test2.mp4"},
				Location:    "Leeds, West Yorkshire, United Kingdom",
				Author:      "https://jay.funabashi.co.uk/",
				Published:   "2018-01-28T00:00:00Z",
				Updated:     "2018-01-29T00:00:00Z",
				Rsvp:        "https://example.com/rsvp1",
				LikeOf:      []string{"https://example.com/like1"},
				BookmarkOf:  []string{"https://example.com/bookmark1"},
				RepostOf:    []string{"https://example.com/repost1"},
				Syndication: []string{"https://example.com/syndicate1"},
				InReplyTo:   []string{"https://example.com/reply1"},
				Comment:     []string{"https://example.com/comment1"},
				Archive:     "201801",
			},
		},
	}

	for _, tc := range tt {
		is := is.NewRelaxed(t)
		result, err := mf2.MfFromJson(tc.inputJSON)
		if err != nil {
			t.Logf("ER ::: %s", err.Error())
		}
		t.Logf("TN ::: %s", tc.name)
		t.Logf("RS ::: %#v", result.ToView())
		t.Logf("EX ::: %#v", tc.expected)
		is.Equal(tc.expected, result.ToView())
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
			inputJSON:   `{"type": ["h-entry"], "properties": {"content": ["hell"]}}`,
			expectedMF2: mf2.MicroFormat{Type: []string{"h-entry"}, Properties: createProperty("content", "hell")},
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
