package tts

import (
	"fmt"
	"testing"
)

func _TestSentence(t *testing.T) {
	content := "【专家确认网传乌龟为外来物种建议不要放生】近日网传图文称，河南宝丰县村民王财犁地犁出大乌龟。看了网上照片，郑州林业局野生动物救护中心主任董朝伟确认，这是原产于美洲的鳄鱼龟，对本地生物危害大。他建议不要放生。若当地无救护条件，建议杀死吃掉，它的肉很鲜美。郑州晚报。"
	sentences := SplitSentence(content, 300)
	for _, s := range sentences {
		fmt.Printf("sentence:%s\n", s)
	}

}
