package model

import (
	"errors"
	"github.com/XMatrixStudio/BlogReaper/graphql"
	"github.com/boltdb/bolt"
	"github.com/globalsign/mgo/bson"
	"sort"
	"strings"
	"time"
)

type PublicModel struct {
	*Model
}

type PublicFeed struct {
	ID         bson.ObjectId `bson:"id"`         // 订阅源的ID
	URL        string        `bson:"url"`        // 订阅源的URL
	Title      string        `bson:"title"`      // 订阅源的标题
	Subtitle   string        `bson:"subtitle"`   // 订阅源的子标题
	Articles   []string      `bson:"articles"`   // 订阅源包括的文章URL
	Follow     int64         `bson:"follow"`     // 订阅数量
	UpdateDate int64         `bson:"updateDate"` // 更新时间，如果超过12小时就更新，或强制更新
}

type PublicArticle struct {
	URL        string   `bson:"url"`
	FeedURL    string   `bson:"feedUrl"`
	Title      string   `bson:"title"`
	Published  string   `bson:"published"`
	Updated    string   `bson:"updated"`
	Content    string   `bson:"content"`
	Summary    string   `bson:"summary"`
	PictureURL string   `bson:"pictureUrl"`
	Categories []string `bson:"categories"`
	Read       int64    `bson:"read"`
}

type PopularArticles struct {
	UpdateDate int64             `bson:"updateDate"` // 更新时间，如果超过12小时就更新，或强制更新
	Articles   []graphql.Article `bson:"articles"`
}

func (m *PublicModel) AddPublicFeed(url, title, subtitle string, articles []string) (publicFeed PublicFeed, err error) {
	return publicFeed, m.Update(func(b *bolt.Bucket) error {
		fb, err := b.CreateBucketIfNotExists([]byte("feed"))
		if err != nil {
			return err
		}
		ufb, err := b.CreateBucketIfNotExists([]byte("key_url_value_id"))
		if err != nil {
			return err
		}
		publicFeed = PublicFeed{
			ID:         bson.NewObjectId(),
			URL:        url,
			Title:      title,
			Subtitle:   subtitle,
			Articles:   articles,
			Follow:     0,
			UpdateDate: time.Now().Unix(),
		}
		bytes, err := bson.Marshal(&publicFeed)
		if err != nil {
			return err
		}
		err = fb.Put([]byte(publicFeed.ID.Hex()), bytes)
		if err != nil {
			return err
		}
		return ufb.Put([]byte(url), []byte(publicFeed.ID.Hex()))
	})
}

func (m *PublicModel) UpdatePublicFeed(id, title, subtitle string, articles []string) (publicFeed PublicFeed, err error) {
	return publicFeed, m.Update(func(b *bolt.Bucket) error {
		fb, err := b.CreateBucketIfNotExists([]byte("feed"))
		if err != nil {
			return err
		}
		bytes := fb.Get([]byte(id))
		err = bson.Unmarshal(bytes, &publicFeed)
		if err != nil {
			return err
		}
		publicFeed.Title = title
		publicFeed.Subtitle = subtitle
		publicFeed.Articles = articles
		publicFeed.UpdateDate = time.Now().Unix()
		bytes, err = bson.Marshal(&publicFeed)
		if err != nil {
			return err
		}
		return fb.Put([]byte(publicFeed.ID.Hex()), bytes)
	})
}

func (m *PublicModel) AddOrUpdatePublicArticles(url string, articles []PublicArticle) (err error) {
	return m.Update(func(b *bolt.Bucket) error {
		ab, err := b.CreateBucketIfNotExists([]byte("article"))
		if err != nil {
			return err
		}
		fb, err := b.CreateBucketIfNotExists([]byte("feed"))
		if err != nil {
			return err
		}
		bytes := fb.Get([]byte(url))
		if bytes == nil {
			for _, v := range articles {
				bytes, err := bson.Marshal(&v)
				if err != nil {
					return err
				}
				err = ab.Put([]byte(v.URL), bytes)
				if err != nil {
					return err
				}
			}
		} else {
			publicFeed := PublicFeed{}
			urlReadNum := make(map[string]int64)
			err = bson.Unmarshal(bytes, &publicFeed)
			if err != nil {
				return err
			}
			for _, v := range publicFeed.Articles {
				bytes = ab.Get([]byte(v))
				if bytes != nil {
					article := PublicArticle{}
					err = bson.Unmarshal(bytes, &article)
					if err != nil {
						return err
					}
					urlReadNum[v] = article.Read
					ab.Delete([]byte(v))
				}
			}
			for _, v := range articles {
				v.Read = urlReadNum[v.URL]
				bytes, err = bson.Marshal(&v)
				if err != nil {
					return err
				}
				err = ab.Put([]byte(v.URL), bytes)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (m *PublicModel) GetPublicFeedByID(id string) (publicFeed PublicFeed, err error) {
	return publicFeed, m.View(func(b *bolt.Bucket) error {
		fb := b.Bucket([]byte("feed"))
		if fb == nil {
			return errors.New("not_found")
		}
		bytes := fb.Get([]byte(id))
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &publicFeed)
	})
}

func (m *PublicModel) GetPublicFeedByURL(url string) (publicFeed PublicFeed, err error) {
	return publicFeed, m.View(func(b *bolt.Bucket) error {
		fb := b.Bucket([]byte("feed"))
		if fb == nil {
			return errors.New("not_found")
		}
		ufb := b.Bucket([]byte("key_url_value_id"))
		if ufb == nil {
			return errors.New("not_found")
		}
		bytes := ufb.Get([]byte(url))
		if bytes == nil {
			return errors.New("not_found")
		}
		bytes = fb.Get(bytes)
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &publicFeed)
	})
}

func (m *PublicModel) GetPublicFeedsByKeyword(keyword string) (publicFeeds []PublicFeed, err error) {
	return publicFeeds, m.View(func(b *bolt.Bucket) error {
		fb := b.Bucket([]byte("feed"))
		if fb == nil {
			return errors.New("not_found")
		}
		return fb.ForEach(func(k, v []byte) error {
			if string(k) == "key_url_value_id" {
				return nil
			}
			publicFeed := PublicFeed{}
			err = bson.Unmarshal(v, &publicFeed)
			if err != nil {
				return err
			}
			if strings.Contains(publicFeed.Title, keyword) || strings.Contains(publicFeed.Subtitle, keyword) {
				publicFeeds = append(publicFeeds, publicFeed)
			}
			return nil
		})
	})
}

func (m *PublicModel) GetPublicFeedsSortedByFollow() (publicFeeds []PublicFeed, err error) {
	return publicFeeds, m.View(func(b *bolt.Bucket) error {
		fb := b.Bucket([]byte("feed"))
		if fb == nil {
			return errors.New("not_found")
		}
		err = fb.ForEach(func(k, v []byte) error {
			if string(k) == "key_url_value_id" {
				return nil
			}
			publicFeed := PublicFeed{}
			err = bson.Unmarshal(v, &publicFeed)
			if err != nil {
				return err
			}
			publicFeeds = append(publicFeeds, publicFeed)
			return nil
		})
		if err != nil {
			return err
		}
		sort.Slice(publicFeeds, func(i, j int) bool {
			return publicFeeds[i].Follow > publicFeeds[j].Follow
		})
		return nil
	})
}

func (m *PublicModel) GetPublicArticleByURL(url string) (article PublicArticle, err error) {
	return article, m.View(func(b *bolt.Bucket) error {
		ab := b.Bucket([]byte("article"))
		if ab == nil {
			return errors.New("not_found")
		}
		bytes := ab.Get([]byte(url))
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &article)
	})
}

func (m *PublicModel) IncreasePublicFeedFollow(id string) (err error) {
	return m.Update(func(b *bolt.Bucket) error {
		fb, err := b.CreateBucketIfNotExists([]byte("feed"))
		if err != nil {
			return err
		}
		bytes := fb.Get([]byte(id))
		if bytes == nil {
			return errors.New("not_found")
		}
		publicFeed := PublicFeed{}
		err = bson.Unmarshal(bytes, &publicFeed)
		if err != nil {
			return err
		}
		publicFeed.Follow++
		bytes, err = bson.Marshal(&publicFeed)
		if err != nil {
			return err
		}
		return fb.Put([]byte(id), bytes)
	})
}

func (m *PublicModel) DecreasePublicFeedFollow(id string) (err error) {
	return m.Update(func(b *bolt.Bucket) error {
		fb, err := b.CreateBucketIfNotExists([]byte("feed"))
		if err != nil {
			return err
		}
		bytes := fb.Get([]byte(id))
		if bytes == nil {
			return errors.New("not_found")
		}
		publicFeed := PublicFeed{}
		err = bson.Unmarshal(bytes, &publicFeed)
		if err != nil {
			return err
		}
		if publicFeed.Follow > 0 {
			publicFeed.Follow--
		} else {
			return errors.New("already_zero")
		}
		bytes, err = bson.Marshal(&publicFeed)
		if err != nil {
			return err
		}
		err = fb.Delete([]byte(id))
		if err != nil {
			return err
		}
		return fb.Put([]byte(id), bytes)
	})
}

func (m *PublicModel) GetPopularArticles() (articles PopularArticles, err error) {
	return articles, m.View(func(b *bolt.Bucket) error {
		bytes := b.Get([]byte("popularArticles"))
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &articles)
	})
}

func (m *PublicModel) UpdatePopularArticles(publicArticles []graphql.Article) (articles PopularArticles, err error) {
	return articles, m.Update(func(b *bolt.Bucket) error {
		articles = PopularArticles{
			UpdateDate: time.Now().Unix(),
			Articles:   publicArticles,
		}
		bytes, err := bson.Marshal(&articles)
		if err != nil {
			return err
		}
		return b.Put([]byte("popularArticles"), bytes)
	})
}
