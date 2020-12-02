package model

import (
	"encoding/json"
	"testing"
)

func TestSchemaJSON(t *testing.T) {

	var schema = &JobSchema{
		EC2: &EC2ProfileSchema{
			Platform:   []string{"cpu", "gpu"},
			PriceModel: []string{"normal", "spot"},
		},
		FFmpeg: &FFmpegProfileSchema{
			Codec:       []string{"h264", "h265"},
			OriginCodec: []string{"h264", "h265"},
			Scale:       []string{"1080p", "720p", "480p", "360p", "240p"},
			Bitrate:     []int{1, 15},
			Profile:     []string{"quality", "latency"},
			Priority:    []int{0, 100},
		},
	}

	data, err := json.Marshal(schema)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(data))

}
