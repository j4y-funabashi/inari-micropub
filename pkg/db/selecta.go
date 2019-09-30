package db

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
)

func NewSelecta(db *sql.DB) Selecta {
	return Selecta{
		db: db,
	}
}

type ArchiveLinkYear struct {
	Year  string `json:"year"`
	Count int    `json:"count"`
}

type ArchiveLinkMonth struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type Selecta struct {
	db *sql.DB
}

func (s Selecta) SelectMediaYearList() ([]ArchiveLinkYear, error) {

	list := []ArchiveLinkYear{}

	rows, err := s.db.Query(
		`SELECT year,count(*) as count FROM media GROUP BY year ORDER BY sort_key DESC `,
	)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		item := ArchiveLinkYear{}
		err := rows.Scan(&item.Year, &item.Count)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (s Selecta) SelectMonthList(year string) ([]ArchiveLinkMonth, error) {

	list := []ArchiveLinkMonth{}

	if year == "" {
		return list, nil
	}

	rows, err := s.db.Query(
		`SELECT month,count(*) as count FROM posts WHERE year = $1 GROUP BY month ORDER BY sort_key DESC `,
		year,
	)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		item := ArchiveLinkMonth{}
		err := rows.Scan(&item.Month, &item.Count)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (s Selecta) SelectMediaMonthList(year string) ([]ArchiveLinkMonth, error) {

	list := []ArchiveLinkMonth{}

	if year == "" {
		return list, nil
	}

	rows, err := s.db.Query(
		`SELECT month,count(*) as count FROM media WHERE year = $1 GROUP BY month ORDER BY sort_key DESC `,
		year,
	)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		item := ArchiveLinkMonth{}
		err := rows.Scan(&item.Month, &item.Count)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (s Selecta) SelectYearList() ([]ArchiveLinkYear, error) {

	list := []ArchiveLinkYear{}

	rows, err := s.db.Query(
		`SELECT year,count(*) as count FROM posts GROUP BY year ORDER BY sort_key DESC `,
	)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		item := ArchiveLinkYear{}
		err := rows.Scan(&item.Year, &item.Count)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (s Selecta) SelectPostList(limit int, afterKey string) (mf2.PostList, error) {

	postList := mf2.PostList{
		Paging: &mf2.ListPaging{},
	}

	count, err := s.fetchPostCount(afterKey, limit)
	if err != nil {
		return postList, err
	}

	postList, err = s.fetchPostList(afterKey, count, limit)
	return postList, err
}

func (s Selecta) SelectPostByURL(uid string) (mf2.MicroFormat, error) {

	rows, err := s.db.Query(
		`SELECT data FROM posts WHERE id = $1`,
		uid,
	)
	if err != nil {
		return mf2.MicroFormat{}, err
	}

	defer rows.Close()

	var mfJSON string
	mf := mf2.MicroFormat{}

	for rows.Next() {
		err := rows.Scan(&mfJSON)
		if err != nil {
			return mf2.MicroFormat{}, err
		}
		err = json.NewDecoder(strings.NewReader(mfJSON)).Decode(&mf)
		if err != nil {
			return mf2.MicroFormat{}, err
		}
	}
	return mf, nil
}

func (s Selecta) SelectMediaByURL(uid string) (mf2.MediaMetadata, error) {

	rows, err := s.db.Query(
		`SELECT data FROM media WHERE id = $1`,
		uid,
	)
	if err != nil {
		return mf2.MediaMetadata{}, err
	}

	defer rows.Close()

	var mfJSON string
	mf := mf2.MediaMetadata{}

	for rows.Next() {
		err := rows.Scan(&mfJSON)
		if err != nil {
			return mf2.MediaMetadata{}, err
		}
		err = json.NewDecoder(strings.NewReader(mfJSON)).Decode(&mf)
		if err != nil {
			return mf2.MediaMetadata{}, err
		}
	}
	return mf, nil
}

func (s Selecta) SelectMediaMonth(year, month string) (mf2.MediaList, error) {

	return s.fetchMediaMonth(year, month)
}

func (s Selecta) SelectMediaList(limit int, afterKey string) (mf2.MediaList, error) {

	postList := mf2.MediaList{
		Paging: &mf2.ListPaging{},
	}

	count, err := s.fetchMediaCount(afterKey, limit)
	if err != nil {
		return postList, err
	}
	postList, err = s.fetchMediaList(afterKey, count, limit)
	return postList, err
}

func rowsToPostList(rows *sql.Rows, count, limit int) (mf2.PostList, error) {
	postList := mf2.PostList{
		Paging: &mf2.ListPaging{},
	}

	defer rows.Close()
	for rows.Next() {
		var mfJSON string
		var sortKey string
		mf := mf2.MicroFormat{}
		err := rows.Scan(&mfJSON, &sortKey)
		if err != nil {
			return postList, err
		}
		err = json.NewDecoder(strings.NewReader(mfJSON)).Decode(&mf)
		if err != nil {
			return postList, err
		}
		postList.Add(mf)
		if count > limit {
			paging := mf2.ListPaging{
				After: sortKey,
			}
			postList.Paging = &paging
		}
	}

	return postList, nil
}

func rowsToMediaList(rows *sql.Rows, count, limit int) (mf2.MediaList, error) {

	postList := mf2.MediaList{
		Paging: &mf2.ListPaging{},
	}

	defer rows.Close()
	for rows.Next() {
		var mfJSON string
		var sortKey string
		var isPublished string
		err := rows.Scan(&mfJSON, &sortKey, &isPublished)
		if err != nil {
			return postList, err
		}
		mf := mf2.MediaMetadata{}
		err = json.NewDecoder(strings.NewReader(mfJSON)).Decode(&mf)
		if err != nil {
			return postList, err
		}

		if len(isPublished) > 0 && isPublished != "0" {
			mf.IsPublished = true
		}

		postList.Add(mf)
		if count > limit {
			paging := mf2.ListPaging{
				After: sortKey,
			}
			postList.Paging = &paging
		}
	}

	return postList, nil
}

func (s Selecta) fetchPostList(afterKey string, count, limit int) (mf2.PostList, error) {
	postList := mf2.PostList{
		Paging: &mf2.ListPaging{},
	}

	if len(afterKey) > 0 {
		rows, err := s.db.Query(
			`SELECT data, sort_key FROM posts WHERE sort_key < $1 ORDER BY sort_key DESC LIMIT $2`,
			afterKey,
			limit,
		)
		if err != nil {
			return postList, err
		}
		postList, err = rowsToPostList(rows, count, limit)
		return postList, err
	}

	rows, err := s.db.Query(
		`SELECT data, sort_key FROM posts ORDER BY sort_key DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return postList, err
	}
	postList, err = rowsToPostList(rows, count, limit)
	return postList, err
}

func (s Selecta) fetchMediaMonth(year, month string) (mf2.MediaList, error) {
	postList := mf2.MediaList{
		Paging: &mf2.ListPaging{},
	}

	rows, err := s.db.Query(
		`SELECT data, sort_key, COALESCE(media_published.id, 0)
FROM media
LEFT JOIN media_published ON media.id = media_published.id
WHERE media.year = $1 AND media.month = $2
ORDER BY sort_key DESC;`,
		year,
		month,
	)
	if err != nil {
		return postList, err
	}
	postList, err = rowsToMediaList(rows, 0, 0)
	return postList, err
}

func (s Selecta) fetchMediaList(afterKey string, count, limit int) (mf2.MediaList, error) {
	postList := mf2.MediaList{
		Paging: &mf2.ListPaging{},
	}

	if len(afterKey) > 0 {
		rows, err := s.db.Query(
			`SELECT data, sort_key, COALESCE(media_published.id, 0)
FROM media
LEFT JOIN media_published ON media.id = media_published.id
WHERE sort_key < $1 ORDER BY sort_key DESC LIMIT $2;`,
			afterKey,
			limit,
		)
		if err != nil {
			return postList, err
		}
		postList, err := rowsToMediaList(rows, count, limit)
		return postList, err
	}

	rows, err := s.db.Query(
		`SELECT data, sort_key, COALESCE(media_published.id, 0)
FROM media
LEFT JOIN media_published ON media.id = media_published.id
ORDER BY sort_key DESC LIMIT $1;`,
		limit,
	)
	if err != nil {
		return postList, err
	}
	postList, err = rowsToMediaList(rows, count, limit)
	return postList, err
}

func (s Selecta) fetchPostCount(afterKey string, limit int) (int, error) {
	var count int
	if len(afterKey) > 0 {
		row := s.db.QueryRow(
			`SELECT count(sort_key) FROM posts WHERE sort_key < $1 LIMIT $2`,
			afterKey,
			limit+1,
		)
		err := row.Scan(&count)
		return count, err
	}
	row := s.db.QueryRow(
		`SELECT count(sort_key) FROM posts LIMIT $1`,
		limit+1,
	)
	err := row.Scan(&count)
	return count, err
}

func (s Selecta) fetchMediaCount(afterKey string, limit int) (int, error) {
	var count int
	if len(afterKey) > 0 {
		row := s.db.QueryRow(
			`SELECT count(*) FROM media WHERE sort_key < $1 ORDER BY sort_key DESC LIMIT $2`,
			afterKey,
			limit+1,
		)
		err := row.Scan(&count)
		return count, err
	}
	row := s.db.QueryRow(
		`SELECT count(*) FROM media ORDER BY sort_key DESC LIMIT $1`,
		limit+1,
	)
	err := row.Scan(&count)
	return count, err
}
