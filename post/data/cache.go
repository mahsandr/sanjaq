package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

const postVisitSet = "postvisits"

func (c *conn) CountPostVisits(postIDs []uint64) error {
	var err error
	for _, postID := range postIDs {
		err = c.redisClient.ZIncr(context.Background(), postVisitSet,
			&redis.Z{Score: 1, Member: postID}).Err()
		if err != nil {
			return err
		}
	}
	c.cacheTop10()
	return nil
}

func (c *conn) DeleteCachedPost(postID uint64) {
	if result, err := c.redisClient.ZRem(context.Background(), postVisitSet, &redis.Z{Member: postID}).Result(); result != 1 || err != nil {
		return
	}
	c.cacheTop10()
}

const cachedPosts = "cachedposts"

func (c *conn) cacheTop10() error {
	ctx := context.Background()
	top10IDs := c.redisClient.ZRevRange(ctx, postVisitSet, 0, 9).Val()
	// select top10 ids
	c.redisClient.Del(ctx, c.redisClient.Keys(ctx, "post:*").Val()...).Err()

	rows, err := c.DBConn().Query(fmt.Sprintf(selectQuery,
		fmt.Sprintf(fileterPostQuery, strings.Join(top10IDs, ","))), 10, 0)
	if err != nil {
		return err
	}
	var (
		title     string
		body      string
		createdAt int64
		postID    string
	)
	for rows.Next() {
		rows.Scan(&postID, &title, &body, &createdAt)
		if err = c.redisClient.HMSet(ctx, "post:"+postID, "title", title, "body", body, "created_at", createdAt).Err(); err != nil {
			return err
		}
	}
	return nil
}
