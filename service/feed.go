package service

import (
	"errors"
	"github.com/XMatrixStudio/BlogReaper/graphql"
	"github.com/XMatrixStudio/BlogReaper/model"
	"sort"
)

type FeedService interface {
	GetModel() *model.FeedModel
	AddFeed(userID, id, categoryID string) (feed graphql.Feed, err error)
	GetFeedsByCategoryID(userID, categoryID string) (feeds []graphql.Feed, err error)
	GetLaterArticles(userID string, page, numPerPage *int) (articles []graphql.Article, err error)
	EditFeed(userID, feedID string, title *string, categoryIDs []string) (success bool, err error)
	RemoveFeed(userID, feedID string) (success bool, err error)
	EditArticle(userID, feedID, url string, read, later *bool) (success bool, err error)
}

type feedService struct {
	Model   *model.FeedModel
	Service *Service
}

func NewFeedService(s *Service, m *model.FeedModel) FeedService {
	return &feedService{
		Model:   m,
		Service: s,
	}
}

func (s *feedService) GetModel() *model.FeedModel {
	return s.Model
}

func (s *feedService) AddFeed(userID, id, categoryID string) (feed graphql.Feed, err error) {
	_, err = s.Service.Category.GetModel().GetCategoryById(userID, categoryID)
	if err != nil {
		return feed, errors.New("invalid_category")
	}
	feed, err = s.Service.Public.GetPublicFeedByID(id)
	if err != nil {
		return feed, errors.New("invalid_id")
	}
	var articlesUrl []string
	for _, v := range feed.Articles {
		articlesUrl = append(articlesUrl, v.URL)
	}
	_, err = s.Model.GetFeedByPublicID(userID, id)
	if err == nil || err.Error() != "not_found" {
		return feed, errors.New("repeat_feed")
	}
	privateFeed, err := s.Model.AddFeed(userID, id, feed.URL, feed.Title, categoryID, articlesUrl)
	if err != nil {
		return
	}
	feed.ID = privateFeed.ID.Hex()
	feed.Title = privateFeed.Title
	err = s.Service.Public.GetModel().IncreasePublicFeedFollow(id)
	return
}

func (s *feedService) GetFeedsByCategoryID(userID, categoryID string) (feeds []graphql.Feed, err error) {
	privateFeeds, err := s.Model.GetFeedsByCategoryID(userID, categoryID)
	if err != nil {
		return
	}
	for _, v := range privateFeeds {
		feed, err := s.Service.Public.GetPublicFeedByID(v.PublicID.Hex())
		if err != nil {
			return feeds, err
		}
		for k := range feed.Articles {
			feed.Articles[k].FeedID = v.ID.Hex()
		}
		for _, priv := range v.Articles {
			var flag bool
			for pubk := range feed.Articles {
				if priv.URL == feed.Articles[pubk].URL {
					feed.Articles[pubk].Later = priv.Later
					feed.Articles[pubk].Read = priv.Read
					flag = true
					break
				}
			}
			if flag == false && priv.Later {
				feed.Articles = append(feed.Articles, graphql.Article{
					URL:        priv.URL,
					Title:      priv.Content.Title,
					Published:  priv.Content.Published,
					Updated:    priv.Content.Updated,
					Content:    priv.Content.Content,
					Summary:    priv.Content.Summary,
					Categories: priv.Content.Categories,
					Read:       priv.Read,
					Later:      priv.Later,
					FeedID:     v.ID.Hex(),
					FeedTitle:  feed.Title,
				})
			}
		}
		sort.Slice(feed.Articles, func(i, j int) bool {
			return feed.Articles[i].Published >= feed.Articles[j].Published
		})
		var articles []model.Article
		for _, av := range feed.Articles {
			articles = append(articles, model.Article{
				URL:     av.URL,
				Read:    av.Read,
				Later:   av.Later,
				Content: nil,
			})
		}
		err = s.Model.UpdateArticles(userID, v.ID.Hex(), articles)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, graphql.Feed{
			ID:             v.ID.Hex(),
			PublicID:       feed.PublicID,
			URL:            v.URL,
			Title:          v.Title,
			Subtitle:       feed.Subtitle,
			Follow:         feed.Follow,
			ArticlesNumber: len(feed.Articles),
			Articles:       feed.Articles,
		})
	}
	return
}

func (s *feedService) GetLaterArticles(userID string, page, numPerPage *int) (articles []graphql.Article, err error) {
	privateArticles, err := s.Model.GetLaterArticle(userID)
	if err != nil && err.Error() == "not_found" {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var start, end int
	if page == nil {
		start = 0
		end = len(privateArticles)
	} else {
		start = (*page - 1) * (*numPerPage)
		end = (*page-1)*(*numPerPage) + *numPerPage
	}
	if len(privateArticles) < start {
		return nil, nil
	} else if len(privateArticles) <= end {
		end = len(privateArticles)
	}
	for i := start; i < end; i++ {
		publicFeed, err := s.Service.Public.GetPublicFeedByURL(privateArticles[i].Content.FeedURL)
		if err != nil {
			return nil, err
		}
		feed, err := s.Service.Feed.GetModel().GetFeedByPublicID(userID, publicFeed.PublicID)
		if err != nil {
			return nil, err
		}
		articles = append(articles, graphql.Article{
			URL:        privateArticles[i].URL,
			Title:      privateArticles[i].Content.Title,
			Published:  privateArticles[i].Content.Published,
			Updated:    privateArticles[i].Content.Updated,
			Content:    privateArticles[i].Content.Content,
			Summary:    privateArticles[i].Content.Summary,
			PictureURL: privateArticles[i].Content.PictureURL,
			Categories: privateArticles[i].Content.Categories,
			Read:       privateArticles[i].Read,
			Later:      privateArticles[i].Later,
			FeedID:     feed.ID.Hex(),
			FeedTitle:  feed.Title,
		})
	}
	return articles, nil
}

// 参数为nil表示不修改
func (s *feedService) EditFeed(userID, feedID string, title *string, categoryIDs []string) (success bool, err error) {
	feed, err := s.Model.GetFeedByID(userID, feedID)
	if err != nil {
		return false, err
	}
	if title == nil {
		title = &feed.Title
	}
	for _, categoryID := range categoryIDs {
		_, err := s.Service.Category.GetModel().GetCategoryById(userID, categoryID)
		if err != nil && err.Error() == "not_found" {
			return false, errors.New("invalid_category")
		} else {
			return false, err
		}
	}
	_, err = s.Model.EditFeed(userID, feedID, *title, categoryIDs)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *feedService) RemoveFeed(userID, feedID string) (success bool, err error) {
	feed, err := s.Model.GetFeedByID(userID, feedID)
	if err != nil {
		return false, err
	}
	pid := feed.PublicID
	err = s.Model.RemoveFeed(userID, feedID)
	if err != nil {
		return false, err
	}
	err = s.Service.Public.GetModel().DecreasePublicFeedFollow(pid.Hex())
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *feedService) EditArticle(userID, feedID, url string, read, later *bool) (success bool, err error) {
	article, err := s.Model.GetArticleByURL(userID, feedID, url)
	if err != nil {
		return false, errors.New("invalid_feed_or_url")
	}
	var readBool, laterBool bool
	if read != nil {
		readBool = *read
	} else {
		readBool = article.Read
	}
	if later != nil {
		laterBool = *later
	} else {
		laterBool = article.Later
	}
	publicArticle := model.PublicArticle{}
	if laterBool {
		publicArticle, err = s.Service.Public.GetModel().GetPublicArticleByURL(url)
		if err != nil {
			return false, err
		}
	}
	err = s.Model.EditArticle(userID, feedID, url, readBool, laterBool, publicArticle)
	if err != nil {
		return false, err
	}
	return true, nil
}
