package pictureCache

import (
	"time"
)

type PictureInfo struct {
	//图片大小
	size uint32

	//图片在node中的index 开始位置
	index uint32

	//存储的节点
	node *CacheNode

	//图片存入时间
	saveTime time.Time

	//图片序号
	seq uint64

	//图片唯一标示码
	pictureFlag string

	//消息读取次数
	timeNum uint8
}
