package controllers

import (
	"strconv"
	"video_feed/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type HotVideo struct {
	models.Video
	Username string `json:"username"`
	IsLiked  bool   `json:"is_liked"`
}

// 用于判断此视频是否被点过赞
func fillLikedForHot(videos []HotVideo, userID uint) {
	//查询点赞过的视频id存入切片
	var likedIDs []uint
	models.DB.Model(&models.Like{}).Where("user_id = ?", userID).Pluck("video_id", &likedIDs)
	//构建 map，键为视频 ID，值为 true，用于 O(1) 查找
	likeMap := make(map[uint]bool)
	for _, id := range likedIDs {
		likeMap[id] = true
	}
	for i := range videos {
		videos[i].IsLiked = likeMap[videos[i].ID]
	}
}

func GetHot(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(uint)

	//设置limit默认值
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit > 50 {
		limit = 50
	}

	ctx := c.Request.Context()
	// 从 Redis 获取热度最高的前 limit 个视频 ID（按分数降序）
	ids, err := models.RDB.ZRevRange(ctx, "hot:rank", 0, int64(limit-1)).Result()
	if err != nil || len(ids) == 0 {
		// 热榜为空时，从数据库按点赞数降序查询
		var videos []HotVideo
		models.DB.Table("videos").
			Select("videos.*,users.username").
			Joins("LEFT JOIN users ON users.id = videos.user_id").
			Order("videos.like_count DESC").
			Limit(limit).
			Find(&videos)

		//从数据库查询完，将查询到的视频写入Redis热榜
		if len(videos) > 0 {
			//批量执行
			pipe := models.RDB.Pipeline()
			for _, v := range videos {
				pipe.ZAdd(ctx, "hot:rank", redis.Z{Score: float64(v.LikeCount), Member: v.ID})
			}
			_, _ = pipe.Exec(ctx)
		}
		//判断一下是否点过赞
		fillLikedForHot(videos, userID)
		c.JSON(200, gin.H{"list": videos})
		return
	}

	//从缓存获取，将ids转化成uint类型存入videoIDs
	var videoIDs []uint
	for _, idStr := range ids {
		id, _ := strconv.ParseUint(idStr, 10, 64)
		videoIDs = append(videoIDs, uint(id))
	}

	//查询视频详情
	var results []HotVideo
	models.DB.Table("videos").
		Select("videos.*,users.username").
		Joins("LEFT JOIN users ON users.id = videos.user_id").
		Where("videos.id IN ?", videoIDs).
		Order("videos.like_count DESC").
		Find(&results)

	//按照redis返回的顺序重组
	videoMap := make(map[uint]HotVideo)
	for _, v := range results {
		videoMap[v.ID] = v
	}
	var videos []HotVideo
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			videos = append(videos, v)
		}
	}

	fillLikedForHot(videos, userID)
	c.JSON(200, gin.H{"list": videos})
}
