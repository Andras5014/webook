package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
}

type GormInteractiveDAO struct {
	db *gorm.DB
}

func (g *GormInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"updated_at": now,
				"status":     1,
			}),
		}).Create(&UserLikeBiz{
			Uid:       uid,
			Biz:       biz,
			BizId:     id,
			Status:    1,
			UpdatedAt: now,
			CreatedAt: now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt":   gorm.Expr("`like_cnt` + 1"),
				"updated_at": now,
			}),
		}).Create(&Interactive{
			Biz:       biz,
			BizId:     id,
			LikeCnt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
	})
}

func (g *GormInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("uid=? AND biz_id = ? AND biz=?", uid, id, biz).
			Updates(map[string]interface{}{
				"updated_at": now,
				"status":     0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			Where("biz =? AND biz_id=?", biz, id).
			Updates(map[string]interface{}{
				"like_cnt":   gorm.Expr("`like_cnt` - 1"),
				"updated_at": now,
			}).Error
	})
}

func (g *GormInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt":   gorm.Expr("read_cnt + 1"),
			"updated_at": now,
		}),
	}).Create(&Interactive{
		Biz:       biz,
		BizId:     bizId,
		ReadCnt:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error
}

func NewInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GormInteractiveDAO{db: db}
}

type Interactive struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"uniqueIndex:biz_id_type"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_id_type"`

	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	CreatedAt  int64
	UpdatedAt  int64
}

type UserLikeBiz struct {
	Id        int64  `gorm:"primaryKey,autoIncrement"`
	Uid       int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId     int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz       string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status    int
	CreatedAt int64
	UpdatedAt int64
}

type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这边还是保留了了唯一索引
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	// 收藏夹的ID
	// 收藏夹ID本身有索引
	Cid       int64 `gorm:"index"`
	CreatedAt int64
	UpdatedAt int64
}
