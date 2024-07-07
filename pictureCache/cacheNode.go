package pictureCache

type CacheNode struct {
	//内存块
	buff []byte

	//node大小
	totalSize uint32

	//当前空闲内存块的index
	curIndex uint32

	//存储着的图片序号结合
	picturesSeq []uint64

	//下一个节点
	next *CacheNode
}

func (node *CacheNode) init(totalSize uint32) {
	node.totalSize = totalSize
	node.curIndex = 0
	node.buff = make([]byte, totalSize, totalSize)
}

//保存图片
func (node *CacheNode) savePicture(pictureBuff []byte) (isSave bool, index uint32) {

	isSave = false
	if node.totalSize-node.curIndex >= uint32(len(pictureBuff)) { //剩余内存足够

		index = node.curIndex
		isSave = true
		copy(node.buff[node.curIndex:], pictureBuff)
		node.curIndex += uint32(len(pictureBuff))
	}

	return
}

//获取图片
func (node *CacheNode) getPicture(index, pictureSize uint32) (pictureBuff []byte) {
	pictureBuff = nil
	if index+pictureSize <= node.curIndex {
		pictureBuff = node.buff[index : index+pictureSize]
	}
	return
}

func (node *CacheNode) getPicturesSeq() []uint64 {
	//所存储图片的图片序号
	return node.picturesSeq
}

func (node *CacheNode) setPictureSeq(seq uint64) {

	node.picturesSeq = append(node.picturesSeq, seq)

}

func (node *CacheNode) clear() {

	node.curIndex = 0
	node.picturesSeq = node.picturesSeq[0:0]

	return
}
