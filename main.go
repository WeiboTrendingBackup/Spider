package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robfig/cron"
	"gopkg.in/mgo.v2/bson"
)

type Item struct {
	Name        string `json:"name" bson:"name"`
	CreatedTime int64  `json:"created_time" bson:"created_time"` // 热搜数据保存到数据库的时间
	Type        int    `json:"type" bson:"type"`                 // 热搜数据的类型，目前只有两种： 0：热搜榜 1:要闻榜（标签榜）
	Index       int    `json:"index" bson:"index"`               // 数据的序号，热搜榜第一个数据是置顶数据，有51条。要闻榜50条
}

type Time struct {
	CreatedTime int64 `json:"created_time" bson:"created_time"` // 热搜数据保存到数据库的时间
}

func init() {
	initDB()
}

// 每月一次，自动把mongodb的数据更新到仓库中，年份/月份/1.csv 命名。
func main() {
	c := cron.New()
	c.AddFunc("@every 1m", func() {
		var now = time.Now().Unix()
		topics := tredingItem(now)
		hashTags := TrendingHashtag(now)

		fmt.Println(now)

		_, err := timeCol.InsertOne(ctx, &Time{
			CreatedTime: now,
		})
		if err != nil {
			log.Fatalln("批量插入时间出错：", err.Error())
		} else {
			fmt.Printf("批量插入时间成功\n")
		}

		items := append(topics, hashTags...)

		_, err = recordCol.InsertMany(ctx, items)
		if err != nil {
			log.Fatalln("批量插入数据出错：", err.Error())
		} else {
			fmt.Printf("批量插入数据成功，热搜榜数据条数：%d，要闻榜数据条数：%d", len(topics), len(hashTags))
		}
	})
	c.Start()

	select {}
	// getData()
}

func getData() {
	batch := []Item{}
	recordCol.Find(ctx, bson.M{"type": 0}).Sort("created_time").All(&batch)
	for _, item := range batch {
		fmt.Println(item.CreatedTime, item.Name, item.Index)
	}
	// fmt.Print(batch)

}

func tredingItem(now int64) []*Item {
	var items []*Item
	// Request the HTML page.
	res, err := http.Get("https://s.weibo.com/top/summary?cate=realtimehot")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find("tbody .td-02").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		name := s.Find("a").Text()
		items = append(items, &Item{
			Name:        name,
			CreatedTime: now,
			Type:        0,
			Index:       i,
		})
	})

	return items
}

func TrendingHashtag(now int64) []*Item {
	var items []*Item
	// Request the HTML page.
	res, err := http.Get("https://s.weibo.com/top/summary?cate=socialevent")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find("tbody .td-02").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		name := s.Find("a").Text()
		items = append(items, &Item{
			Name:        name,
			CreatedTime: now,
			Type:        1,
			Index:       i + 1,
		})
	})

	return items
}
