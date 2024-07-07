package pictureCache

import (
	"reflect"
	"testing"
)

func TestPictureCache_GetPicCacheInfo(t *testing.T) {
	type fields struct {
		maxMemory        uint64
		usedMemory       uint64
		nodeIndex        *CacheNode
		pictures         map[uint64]*PictureInfo
		seq              uint64
		operationChannel chan *OperationInfo
		searchPIC        chan *SearchPIC
	}
	tests := []struct {
		name   string
		fields fields
		want   *fastjson.Value
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &PictureCache{
				maxMemory:        tt.fields.maxMemory,
				usedMemory:       tt.fields.usedMemory,
				nodeIndex:        tt.fields.nodeIndex,
				pictures:         tt.fields.pictures,
				seq:              tt.fields.seq,
				operationChannel: tt.fields.operationChannel,
				searchPIC:        tt.fields.searchPIC,
			}
			if got := cache.GetPicCacheInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPicCacheInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
