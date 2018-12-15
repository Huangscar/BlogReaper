package model

import (
	"github.com/boltdb/bolt"
	"github.com/globalsign/mgo/bson"
	"github.com/kataras/iris/core/errors"
	"sort"
)

type FeedModel struct {
	*Model
}

type Feed struct {
	ID         bson.ObjectId   `bson:"id"`         // 订阅源的ID
	PublicID   bson.ObjectId   `bson:"publicId"`   // 订阅源的公共ID
	URL        string          `bson:"url"`        // 订阅源的URL
	Title      string          `bson:"title"`      // 订阅源的标题
	Categories []bson.ObjectId `bson:"categories"` // 订阅源的分类
	Articles   []Article       `bson:"articles"`   // 订阅源包括的文章
}

type Article struct {
	URL     string         `bson:"url"`
	Read    bool           `bson:"read"`
	Later   bool           `bson:"later"`
	Content *PublicArticle `bson:"content"`
}

func (m *FeedModel) AddFeed(userID, publicID, url, title, categoryID string, articlesUrl []string) (feed Feed, err error) {
	return feed, m.Update(func(b *bolt.Bucket) error {
		ub, err := b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		pub, err := ub.CreateBucketIfNotExists([]byte("key_pid_value_id"))
		if err != nil {
			return err
		}
		if pub.Get([]byte(publicID)) != nil {
			return errors.New("repeat_url")
		}
		var articles []Article
		for _, a := range articlesUrl {
			articles = append(articles, Article{
				URL:     a,
				Read:    false,
				Later:   false,
				Content: nil,
			})
		}
		feed = Feed{
			ID:         bson.NewObjectId(),
			PublicID:   bson.ObjectIdHex(publicID),
			URL:        url,
			Title:      title,
			Categories: []bson.ObjectId{bson.ObjectIdHex(categoryID)},
			Articles:   articles,
		}
		bytes, err := bson.Marshal(&feed)
		if err != nil {
			return err
		}
		err = ub.Put([]byte(feed.ID.Hex()), bytes)
		if err != nil {
			return err
		}
		return pub.Put([]byte(publicID), []byte(feed.ID.Hex()))
	})
}

func (m *FeedModel) UpdateArticles(userID, feedID string, articles []Article) (err error) {
	return m.Update(func(b *bolt.Bucket) error {
		ub, err := b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("not_found")
		}
		feed := Feed{}
		err = bson.Unmarshal(bytes, &feed)
		if err != nil {
			return err
		}
		for k := range articles {
			if articles[k].Later {
				for _, oldV := range feed.Articles {
					if oldV.URL == articles[k].URL {
						articles[k].Content = oldV.Content
						break
					}
				}
			}
		}
		feed.Articles = articles
		bytes, err = bson.Marshal(&feed)
		if err != nil {
			return err
		}
		return ub.Put([]byte(feedID), bytes)
	})
}

func (m *FeedModel) GetFeedByID(userID, feedID string) (feed Feed, err error) {
	return feed, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &feed)
	})
}

func (m *FeedModel) GetFeedByPublicID(userID, publicID string) (feed Feed, err error) {
	return feed, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		pub := ub.Bucket([]byte("key_pid_value_id"))
		if pub == nil {
			return errors.New("not_found")
		}
		bytes := pub.Get([]byte(publicID))
		if bytes == nil {
			return errors.New("not_found")
		}
		bytes = ub.Get(bytes)
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &feed)
	})
}

func (m *FeedModel) GetFeedsByCategoryID(userID, categoryID string) (feeds []Feed, err error) {
	return feeds, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		return ub.ForEach(func(k, v []byte) error {
			if string(k) == "key_pid_value_id" {
				return nil
			}
			feed := Feed{}
			err = bson.Unmarshal(v, &feed)
			if err != nil {
				return err
			}
			for _, cid := range feed.Categories {
				if categoryID == cid.Hex() {
					feeds = append(feeds, feed)
					break
				}
			}
			return nil
		})
	})
}

func (m *FeedModel) EditFeed(userID, feedID, title string, categoryIDs []string) (feed Feed, err error) {
	return feed, m.Update(func(b *bolt.Bucket) error {
		ub, err := b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		pub, err := ub.CreateBucketIfNotExists([]byte("key_pid_value_id"))
		if err != nil {
			return err
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("invalid_id")
		}
		oldFeed := Feed{}
		err = bson.Unmarshal(bytes, &oldFeed)
		if err != nil {
			return err
		}
		url := oldFeed.URL
		articles := oldFeed.Articles
		publicId := oldFeed.PublicID
		var categories []bson.ObjectId
		if categoryIDs == nil {
			categories = oldFeed.Categories
		}
		err = pub.Delete([]byte(publicId))
		if err != nil {
			return err
		}
		err = ub.Delete([]byte(feedID))
		if err != nil {
			return err
		}

		if categoryIDs != nil {
			for _, id := range categoryIDs {
				categories = append(categories, bson.ObjectIdHex(id))
			}
		}

		feed := Feed{
			ID:         bson.ObjectIdHex(feedID),
			PublicID:   publicId,
			URL:        url,
			Title:      title,
			Categories: categories,
			Articles:   articles,
		}
		bytes, err = bson.Marshal(&feed)
		if err != nil {
			return err
		}

		err = ub.Put([]byte(feedID), bytes)
		if err != nil {
			return err
		}
		err = pub.Put([]byte(publicId), []byte(feedID))
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *FeedModel) RemoveFeed(userID, feedID string) (err error) {
	return m.Update(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		pub := ub.Bucket([]byte("key_pid_value_id"))
		if pub == nil {
			return errors.New("not_found")
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("not_found")
		}
		feed := Feed{}
		bson.Unmarshal(bytes, &feed)
		pid := feed.PublicID
		err = ub.Delete([]byte(feedID))
		if err != nil {
			return err
		}
		err = pub.Delete([]byte(pid.Hex()))
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *FeedModel) GetArticleByURL(userID, feedID, url string) (article Article, err error) {
	return article, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("not_found")
		}
		feed := Feed{}
		err = bson.Unmarshal(bytes, &feed)
		if err != nil {
			return err
		}
		for k := range feed.Articles {
			if feed.Articles[k].URL == url {
				article = feed.Articles[k]
				break
			}
		}
		return nil
	})
}

func (m *FeedModel) GetLaterArticle(userID string) (articles []Article, err error) {
	return articles, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		err = ub.ForEach(func(k, v []byte) error {
			if string(k) == "key_pid_value_id" {
				return nil
			}
			feed := Feed{}
			err = bson.Unmarshal(v, &feed)
			if err != nil {
				return err
			}
			for _, v := range feed.Articles {
				if v.Later {
					articles = append(articles, v)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Content.Published >= articles[j].Content.Published
		})
		return nil
	})
}

func (m *FeedModel) EditArticle(userID, feedID, url string, read, later bool, article PublicArticle) (err error) {
	return m.Update(func(b *bolt.Bucket) error {
		ub, err := b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("not_found")
		}
		feed := Feed{}
		err = bson.Unmarshal(bytes, &feed)
		if err != nil {
			return err
		}
		for k := range feed.Articles {
			if feed.Articles[k].URL == url {
				feed.Articles[k].Read = read
				feed.Articles[k].Later = later
				if later == true {
					feed.Articles[k].Content = &article
				} else {
					feed.Articles[k].Content = nil
				}
				break
			}
		}
		bytes, err = bson.Marshal(&feed)
		if err != nil {
			return err
		}
		return ub.Put([]byte(feedID), bytes)
	})
}

func (m *FeedModel) GetCategoryByFeedID(userID, feedID string) (categorieIDs []string, err error){
	return categorieIDs, m.View(func(b* bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		bytes := ub.Get([]byte(feedID))
		if bytes == nil {
			return errors.New("not_found")
		}
		feed := Feed{}
		err := bson.Unmarshal(bytes, &feed)
		if err != nil {
			return err
		}
		for _, categoryID := range feed.Categories{
			categorieIDs = append(categorieIDs, categoryID.Hex())
		}
		return nil 
	})
}
