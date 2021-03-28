package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
)

var recordCol *qmgo.Collection
var timeCol *qmgo.Collection
var ctx context.Context
var dbClient *qmgo.Client

var MONGODB_URI string

func loadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("加载 .env 文件失败", err.Error())
	}

	MONGODB_URI = os.Getenv("MONGODB_URI")
}

// 因为此函数依赖log，因此不能放在init函数中执行，需要后于log.go的init函数执行
func initDB() {
	loadConfig()

	ctx = context.Background()
	var err error
	dbClient, err = qmgo.NewClient(ctx, &qmgo.Config{Uri: MONGODB_URI})
	if err != nil {
		log.Fatalln("连接 Mongo 数据库报错", err.Error())
	}

	db := dbClient.Database("weibo")
	recordCol = db.Collection("record")
	timeCol = db.Collection("time")

	// 各个数据的索引key，必须有一个非重复的数据，比如id。
	// Unique: 你的索引条件是否要求唯一。注意：是说整个【Key数组】匹配的结果是否唯一，而不是说单独的key是否唯一。
	// 每次修改索引之后，得手动删掉数据库的collection才能生效。（似乎只删除coll下面的 indexes 文件夹就行？

	// 时间，唯一
	timeCol.CreateOneIndex(ctx, options.IndexModel{Key: []string{"created_time"}, Unique: true, Background: true})
	recordCol.CreateOneIndex(ctx, options.IndexModel{Key: []string{"created_time", "type"}, Unique: false, Background: true})
}
