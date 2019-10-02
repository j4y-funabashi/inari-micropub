package app

import (
	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/sirupsen/logrus"
)

type Server struct {
	logger  *logrus.Logger
	selecta db.Selecta
}

type QueryPostListResponse struct {
	PostList []mf2.MicroFormatView
	AfterKey string
}

func New(
	logger *logrus.Logger,
	selecta db.Selecta,
) Server {
	return Server{
		logger:  logger,
		selecta: selecta,
	}
}

func (s Server) QueryPostList(limit int, after string) (*QueryPostListResponse, error) {
	pl, err := s.selecta.SelectPostList(limit, after)
	if err != nil {
		return nil, err
	}
	postList := []mf2.MicroFormatView{}

	for _, mf2 := range pl.Items {
		postList = append(postList, mf2.ToView())
	}

	return &QueryPostListResponse{
		PostList: postList,
		AfterKey: pl.Paging.After,
	}, nil
}
