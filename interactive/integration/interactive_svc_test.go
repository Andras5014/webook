package integration

import (
	interactivev1 "github.com/Andras5014/gohub/api/proto/gen/interactive/v1"
	"github.com/Andras5014/gohub/interactive/integration/startup"
	"github.com/Andras5014/gohub/interactive/repository/dao"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"testing"
	"time"
)

type InteractiveTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func (s *InteractiveTestSuite) SetupSuite() {

	config := startup.InitConfig()
	l := startup.InitLogger()
	s.db = startup.InitDB(config, l)
	s.rdb = startup.InitRedis(config)
}

func (s *InteractiveTestSuite) TearDownSuite() {
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
	assert.NoError(s.T(), err)
}

func (s *InteractiveTestSuite) TestIncrReadCnt() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64

		wantErr  error
		wantResp *interactivev1.IncrReadCntResponse
	}{
		{
			// DB 和缓存都有数据
			name: "增加成功,db和redis",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					CreatedAt:  6,
					UpdatedAt:  7,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2",
					"read_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdatedAt > 7)
				data.UpdatedAt = 0
				assert.Equal(t, dao.Interactive{
					Id:    1,
					Biz:   "test",
					BizId: 2,
					// +1 之后
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					CreatedAt:  6,
				}, data)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "read_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:      "test",
			bizId:    2,
			wantResp: &interactivev1.IncrReadCntResponse{},
		},
		{
			// DB 有数据，缓存没有数据
			name: "增加成功,db有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					CreatedAt:  6,
					UpdatedAt:  7,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 2).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdatedAt > 7)
				data.UpdatedAt = 0
				assert.Equal(t, dao.Interactive{
					Id:    2,
					Biz:   "test",
					BizId: 3,
					// +1 之后
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					CreatedAt:  6,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:      "test",
			bizId:    3,
			wantResp: &interactivev1.IncrReadCntResponse{},
		},
		{
			name:   "增加成功-都没有",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz_id = ? AND biz = ?", 4, "test").First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdatedAt > 0)
				assert.True(t, data.CreatedAt > 0)
				assert.True(t, data.Id > 0)
				data.UpdatedAt = 0
				data.CreatedAt = 0
				data.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   4,
					ReadCnt: 1,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:      "test",
			bizId:    4,
			wantResp: &interactivev1.IncrReadCntResponse{},
		},
	}

	// 不同于 AsyncSms 服务，我们不需要 mock，所以创建一个就可以
	// 不需要每个测试都创建
	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.IncrReadCnt(context.Background(), &interactivev1.IncrReadCntRequest{
				Biz:   tc.biz,
				BizId: tc.bizId,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestLike() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr error
	}{
		{
			name: "点赞-DB和cache都有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					CreatedAt:  6,
					UpdatedAt:  7,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2",
					"like_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdatedAt > 7)
				data.UpdatedAt = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    6,
					CreatedAt:  6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 2, 123).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.CreatedAt > 0)
				assert.True(t, likeBiz.UpdatedAt > 0)
				likeBiz.Id = 0
				likeBiz.CreatedAt = 0
				likeBiz.UpdatedAt = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  2,
					Uid:    123,
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 2,
			uid:   123,
		},
		{
			name:   "点赞-都没有",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?",
					"test", 3).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdatedAt > 0)
				assert.True(t, data.CreatedAt > 0)
				assert.True(t, data.Id > 0)
				data.UpdatedAt = 0
				data.CreatedAt = 0
				data.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   3,
					LikeCnt: 1,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 3, 124).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.CreatedAt > 0)
				assert.True(t, likeBiz.UpdatedAt > 0)
				likeBiz.Id = 0
				likeBiz.CreatedAt = 0
				likeBiz.UpdatedAt = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  3,
					Uid:    124,
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 3,
			uid:   124,
		},
	}

	svc := startup.InitInteractiveSvc()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.Like(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.NoError(t, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestDislike() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr error
	}{
		{
			name: "取消点赞-DB和cache都有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					CreatedAt:  6,
					UpdatedAt:  7,
				}).Error
				assert.NoError(t, err)
				err = s.db.Create(dao.UserLikeBiz{
					Id:        1,
					Biz:       "test",
					BizId:     2,
					Uid:       123,
					CreatedAt: 6,
					UpdatedAt: 7,
					Status:    1,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2",
					"like_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdatedAt > 7)
				data.UpdatedAt = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    4,
					CreatedAt:  6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("id = ?", 1).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.UpdatedAt > 7)
				likeBiz.UpdatedAt = 0
				assert.Equal(t, dao.UserLikeBiz{
					Id:        1,
					Biz:       "test",
					BizId:     2,
					Uid:       123,
					CreatedAt: 6,
					Status:    0,
				}, likeBiz)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 2, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 2,
			uid:   123,
		},
	}

	svc := startup.InitInteractiveSvc()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.CancelLike(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.NoError(t, err)
			tc.after(t)
		})
	}
}

func TestInteractiveService(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}
