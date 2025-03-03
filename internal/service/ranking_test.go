package service

import (
	"context"
	interactivev1 "github.com/Andras5014/gohub/api/proto/gen/interactive/v1"
	domain2 "github.com/Andras5014/gohub/interactive/domain"
	"github.com/Andras5014/gohub/internal/domain"
	svcmocks "github.com/Andras5014/gohub/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestBatchRankingService_TopN(t *testing.T) {
	const batchSize = 2
	now := time.Now()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (interactivev1.InteractiveServiceClient, ArticleService)

		wantArts []domain.Article
		wantErr  error
	}{
		{
			name: "成功获取",
			mock: func(ctrl *gomock.Controller) (interactivev1.InteractiveServiceClient, ArticleService) {
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 先模拟批量获取数据
				// 先模拟第一批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 2).
					Return([]domain.Article{
						{Id: 1, UpdatedAt: now},
						{Id: 2, UpdatedAt: now},
					}, nil)
				// 模拟第二批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, 2).
					Return([]domain.Article{
						{Id: 3, UpdatedAt: now},
						{Id: 4, UpdatedAt: now},
					}, nil)
				// 模拟第三批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, 2).
					// 没数据了
					Return([]domain.Article{}, nil)

				// 第一批的点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2}).
					Return(map[int64]domain2.Interactive{
						1: {LikeCnt: 1},
						2: {LikeCnt: 2},
					}, nil)

				// 第二批的点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{3, 4}).
					Return(map[int64]domain2.Interactive{
						3: {LikeCnt: 3},
						4: {LikeCnt: 4},
					}, nil)

				// 第三批的点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).
					Return(map[int64]domain2.Interactive{}, nil)

				return intrSvc, artSvc
			},

			wantErr: nil,
			wantArts: []domain.Article{
				{Id: 4, UpdatedAt: now},
				{Id: 3, UpdatedAt: now},
				{Id: 2, UpdatedAt: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			intrSvc, artSvc := tc.mock(ctrl)
			svc := &BatchRankingService{
				intrSvc:   intrSvc,
				artSvc:    artSvc,
				batchSize: batchSize,

				scoreFunc: func(likeCnt int64, utime time.Time) float64 {
					return float64(likeCnt)
				},
			}
			arts, err := svc.topN(context.Background(), 3)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
