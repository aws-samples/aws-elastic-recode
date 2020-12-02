// Copyright 2020 AWS ElasticRecode Solution Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @author Su Wei <suwei007@gmail.com>
package model

var ec2Platform = []string{"cpu", "gpu"}

const EC2PriceModelNormal = "normal"
const EC2PriceModelSpot = "spot"

//ServerName header
const ServerName = "ElasticRecode/Controlplane"

//ServerVersion header
const ServerVersion = "v0.1.4"

const FFMPEG_BITRATE_MIN = 1
const FFMPEG_BITRATE_MAX = 15
const FFMPEG_PRIORITY_MIN = 0
const FFMPEG_PRIORITY_MIX = 100

var ffmpegProfile = []string{"quality", "latency"}

var ffmpegScaleMap = map[string]string{
	"1080p": "1920:1080",
	"720p":  "1280:720",
	"480p":  "800:480",
	"360p":  "640:360",
	"240p":  "480:270",
}

var vmafScaleMap = map[string]string{
	"1080p": "1920x1080",
	"720p":  "1280x720",
	"480p":  "800x480",
	"360p":  "640x360",
	"240p":  "480x270",
}

var ffmpegCPUCodesMap = map[string]string{
	"h264": "libx264",
	"h265": "libx265",
}

var ffmpegGPUCodesMap = map[string]string{
	"h264": "h264_nvenc",
	"h265": "hevc_nvenc",
}

var ffmpegGPUOriginCodesMap = map[string]string{
	"h264": "h264_cuvid",
	"h265": "hevc_cuvid",
}

//JobSchema is mode struct
type JobSchema struct {
	EC2    *EC2ProfileSchema    `json:"ec2,omitempty"`
	FFmpeg *FFmpegProfileSchema `json:"ffmpeg,omitempty"`
}

//EC2ProfileSchema  is model struct
type EC2ProfileSchema struct {
	Platform   []string `json:"platform,omitempty"`
	PriceModel []string `json:"priceModel,omitempty"`
}

//FFmpegProfileSchema   is model struc
type FFmpegProfileSchema struct {
	Codec       []string `json:"codec,omitempty"`
	OriginCodec []string `json:"originCodec,omitempty"`
	Scale       []string `json:"scale,omitempty"`
	Bitrate     []int    `json:"bitrate,omitempty"`
	Profile     []string `json:"profile,omitempty"`
	Priority    []int    `json:"priority,omitempty"`
}

func getMapKey(data map[string]string) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

//NewSchema init metadata
func NewSchema() *JobSchema {

	return &JobSchema{
		EC2: &EC2ProfileSchema{
			Platform:   []string{"cpu", "gpu"},
			PriceModel: []string{"normal", "spot"},
		},
		FFmpeg: &FFmpegProfileSchema{
			Codec:       getMapKey(ffmpegCPUCodesMap),
			OriginCodec: getMapKey(ffmpegCPUCodesMap),
			Scale:       getMapKey(ffmpegScaleMap),
			Bitrate:     []int{1, 15},
			Profile:     []string{"quality", "latency"},
			Priority:    []int{0, 100},
		},
	}

}
