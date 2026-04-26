package controllers

import (
	"context"
	"strconv"
	"video_feed/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 更新热榜分数
func updateHotRank(videoID uint) {
	//查找视频是否存在
	var video models.Video
	if err := models.DB.First(&video, videoID).Error; err != nil {
		return
	}
	//根据视频点赞数增加得分
	ctx := context.Background()
	models.RDB.ZAdd(ctx, "hot:rank", redis.Z{
		Score:  float64(video.LikeCount),
		Member: video.ID,
	})
}

// 定义一个控制器函数 ToggleLike，用于切换点赞状态，接收 Gin 上下文
func ToggleLike(c *gin.Context) {

	//从上下文获取userid
	userID, _ := c.Get("user_id")
	//获取 URL 路径参数 id
	videoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的视频ID"})
		return
	}

	//获取到了，继续执行
	var like models.Like //存储查询到的点赞记录
	//使用 GORM 查询 likes 表中是否存在当前用户对当前视频的点赞记录
	result := models.DB.Where("user_id=? AND video_id=?", userID, videoID).First(&like)
	//判断是否点赞
	//点过赞，取消点赞
	if result.Error == nil {
		models.DB.Delete(&like)
		models.DB.Model(&models.Video{}).Where("id = ?", videoID).
			Update("like_count", gorm.Expr("like_count - 1 "))
		//更新热榜
		updateHotRank(uint(videoID))
		//liked返回给前端,like_count来自数据库
		c.JSON(200, gin.H{"liked": false, "like_count": getLikeCount(videoID)})
	} else {
		//未点赞，增加点赞
		like = models.Like{UserID: userID.(uint), VideoID: uint(videoID)}
		models.DB.Create(&like)
		models.DB.Model(&models.Video{}).Where("id = ?", videoID).
			Update("like_count", gorm.Expr("like_count + 1 "))
		//更新热榜
		updateHotRank(uint(videoID))
		c.JSON(200, gin.H{"liked": true, "like_count": getLikeCount(videoID)})
	}
}

// 定义一个辅助函数,方便获取最新点赞数返回客户端
func getLikeCount(videoID int) int64 {
	var count int64
	models.DB.Model(&models.Video{}).Where("id = ?", videoID).Select("like_count").Scan(&count)
	return count
}
