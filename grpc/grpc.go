package grpc

import (
	"github.com/DEEPIR/internalservice/dedup"
	"github.com/inconshreveable/log15"
)

var (
	DedupClient *dedup.Client
)

func InitDedupClient(addr string) {
	DedupClient = dedup.NewClient(addr)
}

func IsStaticVideo(videoId string, frameId int64, image []byte) (bool, float64, error) {
	if len(image) == 0 {
		return false, 0, nil
	}
	args := &dedup.FrameArgs{
		VideoID: videoId,
		Frame: dedup.Frame{
			ID:      frameId,
			Content: image,
		},
	}

	isStatic, score, err := DedupClient.IsStaticVideo(args)
	if err != nil {
		log15.Error("is static error: %s", err)

		return false, 0, err
	}
	return isStatic, score, nil
}
