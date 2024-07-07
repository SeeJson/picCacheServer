package pictureCache

import (
	"fmt"
	"github.com/valyala/fastjson"
	"strconv"
	"testing"
)

func TestPictureCache_GetPicCacheInfo(t *testing.T) {

	cfjs := &fastjson.Value{}
	cache = &PictureCache{
		maxMemory:        1000,
		usedMemory:       110,
		nodeIndex:        nil,
		pictures:         nil,
		seq:              0,
		operationChannel: nil,
		searchPIC:        nil,
	}
	str := fmt.Sprintf("%0.4fMB", float64(cache.maxMemory)/1024/1024)
	cfjs.Set("maxMemory", fastjson.MustParse("`"+str+"`"))

	str = fmt.Sprintf("%0.4fMB", float64(cache.usedMemory)/1024/1024)
	cfjs.Set("usedMemory", fastjson.MustParse(str))
	cfjs.Set("pictures", fastjson.MustParse(strconv.Itoa(len(cache.pictures))))

	fmt.Println(cfjs.String())
}
