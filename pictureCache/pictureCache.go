package pictureCache

import (
	"bytes"
	"fmt"
	"github.com/SeeJson/picCacheServer/bytespool"
	"github.com/valyala/fastjson"
	"math/rand"
	"strconv"
	"time"
)

const (
	nodeMemory       = 50 * 1024 * 1024 /// 50M 暂时没有一张图片超过10M的情况
	TimeBucke  uint8 = 10               // 缓存存活时间
	TimeNum    uint8 = 10               // 访问次数
)

var cache *PictureCache

// GetPictureCache -
func GetPictureCache() (cache *PictureCache) {
	if cache == nil {
		cache = &PictureCache{}
	}
	return cache
}

// OperationInfo -
type OperationInfo struct {
	//图片序号
	seq uint64

	//图片保存时间
	saveTime string

	//图片唯一标示码
	pictureFlag string

	//存放图片内存
	buff []byte

	//操作结果等待
	result chan bool

	//操作嘛 0x01 0x02
	code uint8
}

// SearchPIC - 获取图片
type SearchPIC struct {
	seq     uint64        /// 图片序号
	picFlag string        /// 图片唯一标识码
	resBuff *bytes.Buffer /// 图片数据
	result  chan bool     /// 结果
}

// PictureCache -
type PictureCache struct {
	//最大内存分配量，10M对齐
	maxMemory uint64

	//已经使用的内存空间
	usedMemory uint64

	//当前指向的cacheNode
	nodeIndex *CacheNode

	//存储着的图片
	pictures map[uint64]*PictureInfo

	//给图片分配的序号,足够大，足够循环使用了
	seq uint64

	//channel
	operationChannel chan *OperationInfo /// 图片操作
	searchPIC        chan *SearchPIC     /// 图片查询
}

// Init -
func (cache *PictureCache) Init(maxMemory uint64) {
	cache.maxMemory = maxMemory
	cache.seq = 0
	cache.usedMemory = 0
	cache.nodeIndex = nil
	cache.operationChannel = make(chan *OperationInfo, 1)
	cache.searchPIC = make(chan *SearchPIC, 1)
	cache.pictures = make(map[uint64]*PictureInfo)

	go func() {
		for {
			select {
			case operationInfo := <-cache.operationChannel:
				cache.picOperator(operationInfo)
			case searchPic := <-cache.searchPIC: /// 查询图片
				cache.searchPICOperator(searchPic)
			}
		}
	}()
}

func (cache *PictureCache) picOperator(p *OperationInfo) {
	/// 新增图片、打印图片
	switch p.code {
	case 0x01: //添加图片
		result := cache.savePicture(p.buff)
		p.seq = cache.pictures[cache.seq].seq
		p.saveTime = fmt.Sprintf("%d", cache.pictures[cache.seq].saveTime.UTC().UnixNano()/1e6)
		p.pictureFlag = cache.pictures[cache.seq].pictureFlag
		p.result <- result

	case 0x02: //获取图片
		result := false
		if !cache.checkShotFiringSafety(p.seq, p.pictureFlag) {
			p.result <- result
			return
		}
		pictureBuff := cache.getPicture(p.seq, p.pictureFlag)
		if pictureBuff != nil {
			p.buff = make([]byte, len(pictureBuff), len(pictureBuff))
			copy(p.buff, pictureBuff)
			result = true
		}
		p.result <- result

	case 0x03: //打印图片信息
		p.buff = append(p.buff, []byte("seq	size	time\n")...)
		for _, value := range cache.pictures {
			//todo
			p.buff = append(p.buff, []byte(fmt.Sprintf("%d	%d	%s\n",
				value.seq,
				value.size,
				fmt.Sprintf("%d", value.saveTime.UTC().UnixNano()/1e6),
			))...)
		}
		p.result <- true
	}
}

func (cache *PictureCache) searchPICOperator(p *SearchPIC) {
	result := false

	if !cache.checkShotFiringSafety(p.seq, p.picFlag) {
		p.result <- result
		return
	}

	picBUFF := cache.getPicture(p.seq, p.picFlag)
	if picBUFF != nil {
		pBytes := bytespool.GetBytes()
		pBytes.Write(picBUFF)
		p.resBuff = pBytes
		result = true
	}
	p.result <- result
}

// SavePicture - 对外接口 保存图片
func (cache *PictureCache) SavePicture(pictureBuff []byte) (bool, uint64, string, string) {
	operationInfo := &OperationInfo{
		0,
		"",
		"",
		pictureBuff,
		make(chan bool),
		0x01,
	}

	cache.operationChannel <- operationInfo
	result := <-operationInfo.result //等待结果

	return result, operationInfo.seq, operationInfo.saveTime, operationInfo.pictureFlag
}

// GetPicture - 获取图片
func (cache *PictureCache) GetPicture(seq uint64, pictureFlag string) (*bytes.Buffer, error) {
	chanSearchPIC := &SearchPIC{seq, pictureFlag, nil, make(chan bool)}
	defer close(chanSearchPIC.result)

	cache.searchPIC <- chanSearchPIC
	result := <-chanSearchPIC.result

	if !result {
		return nil, fmt.Errorf("search no result")
	}

	if chanSearchPIC.resBuff == nil {
		return nil, fmt.Errorf("copy result failure")
	}
	return chanSearchPIC.resBuff, nil
}

// PrintPictureInfo -
func (cache *PictureCache) PrintPictureInfo() []byte {
	operationInfo := &OperationInfo{0, "", "", nil, make(chan bool), 0x03}
	cache.operationChannel <- operationInfo
	<-operationInfo.result //等待结果
	return operationInfo.buff
}

// ClearPicture - 对外接口 删除图片缓存
func (cache *PictureCache) DeletePicture() {
	cache.nodeIndex.clear()
	cache.usedMemory = 0
	cache.pictures = make(map[uint64]*PictureInfo)
}

// ClearPicture - 对外接口 清理缓存
func (cache *PictureCache) ClearPicture() {
	cache.nodeIndex.clear()
	cache.usedMemory = 0
	cache.pictures = make(map[uint64]*PictureInfo)
}

// 生成图片唯一标示码
func (cache *PictureCache) createPictureFlag(now time.Time) string {
	return fmt.Sprintf("%d", rand.New(rand.NewSource(now.Unix())).Uint64())
}

// 保存图片
func (cache *PictureCache) savePicture(pictureBuff []byte) bool {

	//第一次调用
	if cache.nodeIndex == nil {
		cache.nodeIndex = &CacheNode{}
		cache.nodeIndex.init(nodeMemory)
		cache.nodeIndex.next = cache.nodeIndex
		cache.usedMemory += nodeMemory
	}

	now := time.Now()
	//把图片保存在nodeIndex所指向的node
	if isSave, index := cache.nodeIndex.savePicture(pictureBuff); isSave {
		//当前节点足够保存
		cache.seq++

		pictureInfo := &PictureInfo{
			uint32(len(pictureBuff)),
			index,
			cache.nodeIndex,
			now,
			cache.seq,
			cache.createPictureFlag(now),
			0,
		}
		cache.pictures[cache.seq] = pictureInfo

		//
		cache.nodeIndex.setPictureSeq(cache.seq)
		return true
	}

	//当前节点内存空间不够

	//还有内存可分配
	if cache.maxMemory-cache.usedMemory >= nodeMemory {
		node := &CacheNode{}
		node.init(nodeMemory)
		cache.usedMemory += nodeMemory

		//把节点插入到队列中
		node.next = cache.nodeIndex.next
		cache.nodeIndex.next = node
		cache.nodeIndex = node
	} else { //内存已经不可分配，只能清除缓存时间最旧的一个节点 cache.nodeIndex.next
		//获取改节点的图像序号
		picturesSeq := cache.nodeIndex.next.getPicturesSeq()

		//把这些图片从图片map中清除
		for _, seq := range picturesSeq {
			delete(cache.pictures, seq)
		}

		//清空cache.nodeIndex.next节点
		cache.nodeIndex.next.clear()

		cache.nodeIndex = cache.nodeIndex.next
	}

	//再一次把图片保存在nodeIndex所指向的node
	if isSave, index := cache.nodeIndex.savePicture(pictureBuff); isSave {
		//当前节点足够保存
		cache.seq++
		pictureInfo := &PictureInfo{
			uint32(len(pictureBuff)),
			index,
			cache.nodeIndex,
			now,
			cache.seq,
			cache.createPictureFlag(now),
			0,
		}
		cache.pictures[cache.seq] = pictureInfo

		//
		cache.nodeIndex.setPictureSeq(cache.seq)

		return true
	}

	return false
}

// 获取图片
func (cache *PictureCache) getPicture(seq uint64, pictureFlag string) (pictureBuff []byte) {
	pictureBuff = nil
	if pictureInfo, ok := cache.pictures[seq]; ok {
		if pictureInfo.pictureFlag == pictureFlag { //唯一标示码校验
			pictureBuff = pictureInfo.node.getPicture(pictureInfo.index, pictureInfo.size)
			pictureInfo.timeNum++
		}
	}
	return
}

// GetPicCacheInfo - pictureCache详细信息
//func (cache *PictureCache) GetPicCacheInfo1() *fastjson.Value {
//	cfjs := cfastjson.NewCfastjsonObj()
//
//	str := fmt.Sprintf("%0.4fMB", float64(cache.maxMemory)/1024/1024)
//	cfjs.Set("maxMemory", str)
//
//	str = fmt.Sprintf("%0.4fMB", float64(cache.usedMemory)/1024/1024)
//	cfjs.Set("usedMemory", str)
//	cfjs.Set("pictures", len(cache.pictures))
//	return cfjs.GetFastJsonVal()
//}

func (cache *PictureCache) GetPicCacheInfo() *fastjson.Value {
	cfjs := &fastjson.Value{}

	str := fmt.Sprintf("%0.4fMB", float64(cache.maxMemory)/1024/1024)
	cfjs.Set("maxMemory", fastjson.MustParse(str))

	str = fmt.Sprintf("%0.4fMB", float64(cache.usedMemory)/1024/1024)
	cfjs.Set("usedMemory", fastjson.MustParse(str))
	cfjs.Set("pictures", fastjson.MustParse(strconv.Itoa(len(cache.pictures))))
	return cfjs
}

// 防爆破验证 false=消息过期不可访问，true=可正常访问
func (cache *PictureCache) checkShotFiringSafety(seq uint64, pictureFlag string) bool {
	if pictureInfo, ok := cache.pictures[seq]; ok {
		if pictureInfo.pictureFlag == pictureFlag {
			timeBucker := time.Duration(TimeBucke) * time.Minute
			timeNum := TimeNum
			if time.Now().Sub(pictureInfo.saveTime) > timeBucker || pictureInfo.timeNum >= timeNum { //超过准许访问时间或者超过访问次数限制
				return false
			}
			return true
		}
	}
	return false
}
