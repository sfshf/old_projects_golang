package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/invoker/internal/dao"
	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

func generateRootCommentIDs(start, limit, exclude int) []int {
	res := make([]int, 0, limit)
	res = append(res, 0)
	for i := 0; i < limit; i++ {
		if start+i != exclude {
			res = append(res, start+i)
		}
	}
	return res
}

func generateAtWho(src []int, exclude int) []int {
	var res []int
	for _, val := range src {
		if val != exclude {
			res = append(res, val)
		}
	}
	return res
}

func generateInts(limit int) []int {
	var res []int
	for i := 0; i < limit; i++ {
		res = append(res, i)
	}
	return res
}

func main() {
	ctx := context.Background()
	mysqlDNS := os.Getenv("INVOKER_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'INVOKER_MYSQL_DNS'")
	}
	DaoManager := dao.NewManager(gdao.NewOption(mysqlDNS, "invoker_gofakeit", 0, gutil.AppEnvDEV))
	limit := int64(100)
	startId := int64(100000000)
	if err := DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// 1. remove old data
		if err := daoManager.SiteDAO.Table(ctx).Delete(&Site{ID: startId}).Error; err != nil {
			return err
		}
		if err := daoManager.CategoryDAO.Table(ctx).Delete(&Category{}, `site_id=?`, startId).Error; err != nil {
			return err
		}
		if err := daoManager.PostDAO.Table(ctx).Delete(&Post{}, `site_id=?`, startId).Error; err != nil {
			return err
		}
		if err := daoManager.CommentDAO.Table(ctx).Delete(&Comment{}, `site_id=?`, startId).Error; err != nil {
			return err
		}
		if err := daoManager.ThumbupDAO.Table(ctx).Delete(&Thumbup{}, `site_id=?`, startId).Error; err != nil {
			return err
		}
		// 2. add new data
		// site
		if err := daoManager.SiteDAO.Create(ctx, &Site{
			ID:   startId,
			Name: gofakeit.AppName(),
		}); err != nil {
			return err
		}
		// category
		if err := daoManager.CategoryDAO.Create(ctx, &Category{
			ID:     startId,
			SiteID: startId,
			Name:   gofakeit.AppName(),
			Posts:  limit,
		}); err != nil {
			return err
		}
		// posts with comments
		for i := int64(0); i < limit; i++ {
			ts1 := time.Now().UnixMilli()
			postID := startId + i
			if err := daoManager.PostDAO.Create(ctx, &Post{
				ID:         postID,
				SiteID:     startId,
				CategoryID: startId,
				Title:      gofakeit.Sentence(gofakeit.RandomInt([]int{5, 15, 20, 30})),
				PostedAt:   ts1,
				PostedBy:   int64(gofakeit.RandomInt([]int{100000317, 100000334, 100000359, 100000360})),
				Content:    gofakeit.Sentence(gofakeit.RandomInt([]int{20, 50, 100, 150, 200, 300})),
				Image:      gofakeit.RandomString([]string{"https://d2y6ia7j6nkf8t.cloudfront.net/images/bd2d33a542f38f3da4acf12849cd70f86ffed72154a2a59d582a1b1a299ccfaf", "https://d2y6ia7j6nkf8t.cloudfront.net/images/1851575ca6ed6d692430904a79530bd0d573ac7412ce8fabf4e09449e4c64d6a"}),
				State:      dao.PostState_Posted,
				Views:      limit,
				Replies:    limit,
				Thumbups:   0,
				Activity:   ts1,
			}); err != nil {
				return err
			}
			ts2 := time.Now().UnixMilli()
			commentsMap := make(map[int64]*Comment) // ID->object
			rootComments := make([]*Comment, 0)
			for j := int64(0); j < limit; j++ {
				commentID := startId + j + i*limit
				if _, has := commentsMap[commentID]; has {
					continue
				}
				postedBy := int64(gofakeit.RandomInt([]int{100000317, 100000334, 100000359, 100000360}))
				comment := &Comment{
					ID:         commentID,
					SiteID:     startId,
					CategoryID: startId,
					PostID:     postID,
					Content:    gofakeit.Sentence(gofakeit.RandomInt([]int{20, 50, 100, 150, 200})),
					PostedAt:   ts2 + 1000000,
					PostedBy:   postedBy,
					AtWho:      0,
					UpdatedAt:  ts2 + 10000000,
					Replies:    0,
					Thumbups:   0,
				}
				rootCommentID := int64(gofakeit.RandomInt(generateRootCommentIDs(int(startId+i*limit), int(limit), int(commentID))))
				if rootCommentID > 0 {
					// random atWho
					comment.AtWho = int64(gofakeit.RandomInt(generateAtWho([]int{0, 100000317, 100000334, 100000359, 100000360}, int(postedBy))))
					// check rootComment
					rootComment, has := commentsMap[rootCommentID]
					if !has {
						rootComment = &Comment{
							ID:            rootCommentID,
							SiteID:        startId,
							CategoryID:    startId,
							PostID:        postID,
							RootCommentID: 0,
							Content:       gofakeit.Sentence(gofakeit.RandomInt([]int{20, 50, 100, 150, 200})),
							PostedAt:      ts2 - 500000,
							PostedBy:      int64(gofakeit.RandomInt([]int{100000317, 100000334, 100000359, 100000360})),
							AtWho:         0,
							UpdatedAt:     ts2 - 500000,
							Replies:       1,
							Thumbups:      0,
						}
						commentsMap[rootCommentID] = rootComment
						rootComments = append(rootComments, rootComment)
					} else {
						if rootComment.RootCommentID > 0 {
							randRoot := rootComments[gofakeit.RandomInt(generateInts(len(rootComments)))]
							randRoot.Replies = randRoot.Replies + 1
							rootCommentID = randRoot.ID
						} else {
							rootComment.Replies = rootComment.Replies + 1
						}
					}
				}
				comment.RootCommentID = rootCommentID
				commentsMap[commentID] = comment
			}
			comments := make([]*Comment, 0, len(commentsMap))
			for _, comment := range commentsMap {
				comments = append(comments, comment)
			}
			if err := daoManager.CommentDAO.Create(ctx, comments); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
		return
	}
}
