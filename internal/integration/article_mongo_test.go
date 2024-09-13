package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Andras5014/webook/internal/domain"
	"github.com/Andras5014/webook/internal/integration/startup"
	"github.com/Andras5014/webook/internal/repository/dao/article"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ArticleMongoHandlerTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoHandlerTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(context *gin.Context) {
		// 直接设置好
		context.Set("userId", int64(123))
		context.Next()
	})
	s.mdb = startup.InitMongoDB()
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	err = article.InitCollections(s.mdb)
	if err != nil {
		panic(err)
	}
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	hdl := startup.InitArticleHandlerV1(article.NewMongoDBDAO(s.mdb, node))
	hdl.RegisterRoutes(s.server)
}

func (s *ArticleMongoHandlerTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.mdb.Collection("articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoHandlerTestSuite) TestCleanMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.mdb.Collection("articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").
		DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}
func (s *ArticleMongoHandlerTestSuite) TestArticleHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		// 构造请求，直接使用 req
		// 也就是说，我们放弃测试 Bind 的异常分支
		req Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{"author_id", 123}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.CreatedAt > 0)
				assert.True(t, art.UpdatedAt > 0)
				// 我们断定 ID 生成了
				assert.True(t, art.Id > 0)
				// 重置了这些值，因为无法比较
				art.UpdatedAt = 0
				art.CreatedAt = 0
				art.Id = 0
				assert.Equal(t, article.Article{
					Title:    "hello，你好",
					Content:  "随便试试",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},

		{
			// 这个是已经有了，然后修改之后再保存
			name: "更新帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:        2,
					Title:     "我的标题",
					Content:   "我的内容",
					CreatedAt: 456,
					UpdatedAt: 234,
					AuthorId:  123,
					Status:    domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.UpdatedAt > 234)
				art.UpdatedAt = 0
				assert.Equal(t, article.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					// 创建时间没变
					CreatedAt: 456,
					Status:    domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "更新别人的帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:        3,
					Title:     "我的标题",
					Content:   "我的内容",
					CreatedAt: 456,
					UpdatedAt: 234,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, article.Article{
					Id:        3,
					Title:     "我的标题",
					Content:   "我的内容",
					CreatedAt: 456,
					UpdatedAt: 234,
					AuthorId:  789,
					Status:    domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, result.Code)
			// 只能判定有 ID，因为雪花算法你无法确定具体的值
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

func TestArticleMongo(t *testing.T) {
	suite.Run(t, new(ArticleMongoHandlerTestSuite))
}
