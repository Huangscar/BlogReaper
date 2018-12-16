package service

import (
	"encoding/xml"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/XMatrixStudio/BlogReaper/graphql"
	"github.com/XMatrixStudio/BlogReaper/model"
	"io/ioutil"
	"math/rand"
	"net/http"
	stdUrl "net/url"
	"strings"
	"time"
)

type PublicService interface {
	GetModel() *model.PublicModel
	GetPublicFeedByID(id string) (feed graphql.Feed, err error)
	GetPublicFeedByURL(url string) (feed graphql.Feed, err error)
	GetPublicFeedByKeyword(keyword string) (feeds []graphql.Feed, err error)
	GetPopularPublicFeeds(page, numPerPage int) (feeds []graphql.Feed, err error)
	GetPopularPublicArticles(page, numPerPage int) (articles []graphql.Article, err error)
}

type publicService struct {
	Model   *model.PublicModel
	Service *Service
}

func NewPublicService(s *Service, m *model.PublicModel) PublicService {
	return &publicService{
		Model:   m,
		Service: s,
	}
}

func (s *publicService) GetModel() *model.PublicModel {
	return s.Model
}

// 从数据库中获取PublicFeed
func (s *publicService) GetPublicFeedByID(id string) (feed graphql.Feed, err error) {
	publicFeed, err := s.Model.GetPublicFeedByID(id)
	if err != nil {
		return
	}
	feed = graphql.Feed{
		ID:             "",
		PublicID:       publicFeed.ID.Hex(),
		URL:            publicFeed.URL,
		Title:          publicFeed.Title,
		Subtitle:       publicFeed.Subtitle,
		Follow:         int(publicFeed.Follow),
		ArticlesNumber: 0,
		Articles:       nil,
	}
	for _, v := range publicFeed.Articles {
		publicArticle, err := s.Model.GetPublicArticleByURL(v)
		if err != nil {
			return feed, err
		}
		feed.Articles = append(feed.Articles, graphql.Article{
			URL:        publicArticle.URL,
			Title:      publicArticle.Title,
			Published:  publicArticle.Published,
			Updated:    publicArticle.Updated,
			Content:    publicArticle.Content,
			Summary:    publicArticle.Summary,
			PictureURL: publicArticle.PictureURL,
			Categories: publicArticle.Categories,
			Read:       false,
			Later:      false,
			FeedID:     "",
			FeedTitle:  "",
		})
	}
	feed.ArticlesNumber = len(feed.Articles)
	return
}

func (s *publicService) GetPublicFeedByURL(url string) (feed graphql.Feed, err error) {
	u, err := stdUrl.Parse(url)
	if err != nil {
		return feed, errors.New("invalid_url")
	}
	notSchemaUrl := u.Host + u.Path
	publicFeed, err := s.Model.GetPublicFeedByURL(notSchemaUrl)
	if err != nil && err.Error() != "not_found" {
		return
	}
	if err != nil && err.Error() == "not_found" {
		publicFeed, err = s.UpdatePublicFeed("", url)
	} else if time.Now().Unix()-publicFeed.UpdateDate > 60*60*12 {
		publicFeed, err = s.UpdatePublicFeed(publicFeed.ID.Hex(), url)
	}
	if err != nil {
		return
	}
	feed = graphql.Feed{
		ID:             "",
		PublicID:       publicFeed.ID.Hex(),
		URL:            publicFeed.URL,
		Title:          publicFeed.Title,
		Subtitle:       publicFeed.Subtitle,
		Follow:         int(publicFeed.Follow),
		ArticlesNumber: 0,
		Articles:       []graphql.Article{},
	}
	for _, v := range publicFeed.Articles {
		publicArticle, err := s.Model.GetPublicArticleByURL(v)
		if err != nil {
			return feed, err
		}
		feed.Articles = append(feed.Articles, graphql.Article{
			URL:        publicArticle.URL,
			Title:      publicArticle.Title,
			Published:  publicArticle.Published,
			Updated:    publicArticle.Updated,
			Content:    publicArticle.Content,
			Summary:    publicArticle.Summary,
			PictureURL: publicArticle.PictureURL,
			Categories: publicArticle.Categories,
			Read:       false,
			Later:      false,
			FeedID:     "",
			FeedTitle:  feed.Title,
		})
	}
	feed.ArticlesNumber = len(feed.Articles)
	return
}

func (s *publicService) GetPublicFeedByKeyword(keyword string) (feeds []graphql.Feed, err error) {
	publicFeeds, err := s.Model.GetPublicFeedsByKeyword(keyword)
	if err != nil {
		return
	}
	for _, v := range publicFeeds {
		feed := graphql.Feed{
			ID:             "",
			PublicID:       v.ID.Hex(),
			URL:            v.URL,
			Title:          v.Title,
			Subtitle:       v.Subtitle,
			Follow:         int(v.Follow),
			ArticlesNumber: 0,
			Articles:       []graphql.Article{},
		}
		for _, v := range v.Articles {
			publicArticle, err := s.Model.GetPublicArticleByURL(v)
			if err != nil {
				return feeds, err
			}
			feed.Articles = append(feed.Articles, graphql.Article{
				URL:        publicArticle.URL,
				Title:      publicArticle.Title,
				Published:  publicArticle.Published,
				Updated:    publicArticle.Updated,
				Content:    publicArticle.Content,
				Summary:    publicArticle.Summary,
				PictureURL: publicArticle.PictureURL,
				Categories: publicArticle.Categories,
				Read:       false,
				Later:      false,
				FeedID:     "",
				FeedTitle:  feed.Title,
			})
		}
		feed.ArticlesNumber = len(feed.Articles)
		feeds = append(feeds, feed)
	}
	return
}

func (s *publicService) GetPopularPublicFeeds(page, numPerPage int) (feeds []graphql.Feed, err error) {
	publicFeeds, err := s.Model.GetPublicFeedsSortedByFollow()
	if err != nil && err.Error() == "not_found" {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	start := (page - 1) * (numPerPage)
	end := (page-1)*(numPerPage) + numPerPage
	if len(publicFeeds) < start {
		return nil, nil
	} else if len(publicFeeds) <= end {
		end = len(publicFeeds)
	}
	for i := start; i < end; i++ {
		var articles []graphql.Article
		for _, v := range publicFeeds[i].Articles {
			article, err := s.Model.GetPublicArticleByURL(v)
			if err != nil {
				return nil, err
			}
			articles = append(articles, graphql.Article{
				URL:        article.URL,
				Title:      article.Title,
				Published:  article.Published,
				Updated:    article.Updated,
				Content:    article.Content,
				Summary:    article.Summary,
				PictureURL: article.PictureURL,
				Categories: article.Categories,
				Read:       false,
				Later:      false,
				FeedID:     "",
				FeedTitle:  publicFeeds[i].Title,
			})
		}
		feeds = append(feeds, graphql.Feed{
			ID:             "",
			PublicID:       publicFeeds[i].ID.Hex(),
			URL:            publicFeeds[i].URL,
			Title:          publicFeeds[i].Title,
			Subtitle:       publicFeeds[i].Subtitle,
			Follow:         int(publicFeeds[i].Follow),
			ArticlesNumber: len(articles),
			Articles:       articles,
		})
	}
	return feeds, nil
}

func (s *publicService) GetPopularPublicArticles(page, numPerPage int) (articles []graphql.Article, err error) {
	popularArticles, err := s.Model.GetPopularArticles()
	if (err != nil && err.Error() == "not_found") || time.Now().Unix()-popularArticles.UpdateDate > 60*60*12 {
		articles, err = s.UpdatePopularPublicArticles()
		if err != nil {
			articles = popularArticles.Articles
		}
	} else {
		articles = popularArticles.Articles
	}
	start := (page - 1) * (numPerPage)
	end := (page-1)*(numPerPage) + numPerPage
	if len(articles) < start {
		return nil, nil
	} else if len(articles) <= end {
		end = len(articles)
	}
	return articles[start:end], nil
}

func (s *publicService) UpdatePopularPublicArticles() (articles []graphql.Article, err error) {
	feeds, err := s.GetPopularPublicFeeds(1, 100)
	if err != nil {
		return articles, err
	}
	for _, v := range feeds {
		for _, a := range v.Articles {
			articles = append(articles, a)
		}
	}
	dst := make([]graphql.Article, len(articles))
	perm := rand.Perm(len(articles))
	for i, v := range perm {
		dst[v] = articles[i]
	}
	if len(dst) >= 100 {
		articles = dst[:100]
	} else {
		articles = dst
	}
	_, err = s.Model.UpdatePopularArticles(articles)
	return
}

type AtomFeed struct {
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle"`
	Author   AtomAuthor  `xml:"author"`
	Entries  []AtomEntry `xml:"entry"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomEntry struct {
	Title      string         `xml:"title"`
	Link       AtomLink       `xml:"link"`
	Published  string         `xml:"published"`
	Updated    string         `xml:"updated"`
	Content    string         `xml:"content"`
	Summary    string         `xml:"summary"`
	Categories []AtomCategory `xml:"category"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
}

type AtomCategory struct {
	Term string `xml:"term,attr"`
}

type RSSTop struct {
	Channel RSSFeed `xml:"channel"`
}

type RSSFeed struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	Categories  []string `xml:"category"`
}

// 从订阅源拉取数据，更新PublicFeed
func (s *publicService) UpdatePublicFeed(id, url string) (publicFeed model.PublicFeed, err error) {
	u, err := stdUrl.Parse(url)
	if err != nil {
		return publicFeed, errors.New("invalid_url")
	}
	notSchemaUrl := u.Host + u.Path
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return publicFeed, errors.New(res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	atomFeed := AtomFeed{}
	err = xml.Unmarshal(bytes, &atomFeed)
	if len(atomFeed.Title) == 0 && len(atomFeed.Entries) == 0 {
		/* atom解析失败，解析rss*/
		rssTop := RSSTop{}
		err = xml.Unmarshal(bytes, &rssTop)
		if err != nil {
			return publicFeed, err
		}
		rssFeed := rssTop.Channel
		if len(rssFeed.Title) == 0 && len(rssFeed.Items) == 0 {
			return publicFeed, errors.New("invalid_url")
		}
		var articlesUrl []string
		var articles []model.PublicArticle
		for _, v := range rssFeed.Items {
			var categories []string
			for _, vc := range v.Categories {
				categories = append(categories, vc)
			}
			document, err := goquery.NewDocumentFromReader(strings.NewReader(v.Description))
			if err != nil {
				return publicFeed, err
			}
			find := document.Find("img")
			var picture string
			if find == nil || find.First() == nil {
				picture = ""
			} else {
				picture = find.First().AttrOr("src", "")
			}
			arrRune := []rune(document.Text())
			if len(arrRune) > 200 {
				arrRune = arrRune[:200]
			}
			articlesUrl = append(articlesUrl, v.Link)
			articles = append(articles, model.PublicArticle{
				URL:        v.Link,
				FeedURL:    url,
				Title:      v.Title,
				Published:  v.PubDate,
				Updated:    "",
				Content:    v.Description,
				Summary:    string(arrRune),
				PictureURL: picture,
				Categories: categories,
				Read:       0,
			})
		}
		err = s.Model.AddOrUpdatePublicArticles(notSchemaUrl, articles)
		if err != nil {
			return
		}
		if id == "" {
			publicFeed, err = s.Model.AddPublicFeed(notSchemaUrl, rssFeed.Title, "", articlesUrl)
		} else {
			publicFeed, err = s.Model.UpdatePublicFeed(id, rssFeed.Title, "", articlesUrl)
		}
		return
	}
	/* atom解析成功*/
	var articlesUrl []string
	var articles []model.PublicArticle
	for _, v := range atomFeed.Entries {
		var categories []string
		for _, vc := range v.Categories {
			categories = append(categories, vc.Term)
		}
		document, err := goquery.NewDocumentFromReader(strings.NewReader(v.Content))
		if err != nil {
			return publicFeed, err
		}
		find := document.Find("img")
		var picture string
		if find == nil || find.First() == nil {
			picture = ""
		} else {
			picture = find.First().AttrOr("src", "")
		}
		articlesUrl = append(articlesUrl, v.Link.Href)
		articles = append(articles, model.PublicArticle{
			URL:        v.Link.Href,
			FeedURL:    url,
			Title:      v.Title,
			Published:  v.Published,
			Updated:    v.Updated,
			Content:    v.Content,
			Summary:    v.Summary,
			PictureURL: picture,
			Categories: categories,
			Read:       0,
		})
	}
	err = s.Model.AddOrUpdatePublicArticles(notSchemaUrl, articles)
	if err != nil {
		return
	}
	if id == "" {
		publicFeed, err = s.Model.AddPublicFeed(notSchemaUrl, atomFeed.Title, atomFeed.Subtitle, articlesUrl)
	} else {
		publicFeed, err = s.Model.UpdatePublicFeed(id, atomFeed.Title, atomFeed.Subtitle, articlesUrl)
	}
	_, err = s.UpdatePopularPublicArticles()
	return
}
