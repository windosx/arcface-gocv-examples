package main

import (
	"fmt"
	. "github.com/windosx/face-engine/v3"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"strconv"
	"time"
)

var (
	engine *FaceEngine
	window *gocv.Window
	media  *gocv.VideoCapture
	ticker *time.Ticker
)

// 激活SDK
func init() {
	err := Activation("Your AppID", "Your SdkKey")
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error
	// 初始化人脸引擎
	engine, err = NewFaceEngine(DetectModeVideo,
		OrientPriority0,
		16,
		50,
		EnableFaceDetect | EnableAge | EnableGender)
	if err != nil {
		panic(err)
	}
	// 打开视频文件（若是网络摄像头支持RTSP推流，可以使用gocv.OpenVideoCapture("rtsp://<username>:<pwd>@host:port")）
	media, err = gocv.VideoCaptureFile("test.mp4")
	if err != nil {
		panic(err)
	}
	// 整个窗口方便看效果
	window = gocv.NewWindow("Test")
	// 获取视频宽度
	w := media.Get(gocv.VideoCaptureFrameWidth)
	// 获取视频高度
	h := media.Get(gocv.VideoCaptureFrameHeight)
	// 调整窗口大小
	window.ResizeWindow(int(w), int(h))
	// 获取FPS
	fps := media.Get(gocv.VideoCaptureFPS)
	// 获取视频帧数（!!注意，直播流不能根据frames来调整刷新率，获取到帧直接处理即可）
	frames := media.Get(gocv.VideoCaptureFrameCount)
	ticker = time.NewTicker(time.Millisecond * time.Duration(fps))
	for currentFrame := 1; currentFrame <= int(frames); currentFrame++ {
		<-ticker.C
		img := gocv.NewMat()
		media.Read(&img)
		detectFace(engine, &img)
		window.IMShow(img)
		window.WaitKey(1)
		// 图片处理完毕记得关闭以释放内存
		img.Close()
	}
	// 收尾工作
	ticker.Stop()
	media.Close()
	engine.Destroy()
	window.Close()
}

// 虹软开始干活
func detectFace(engine *FaceEngine, img *gocv.Mat) {
	width := img.Cols()
	height := img.Rows()
	faceInfo, err := engine.DetectFaces(width - width % 4, height, ColorFormatBGR24, img.DataPtrUint8())
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if faceInfo.FaceNum > 0 {
		err = engine.Process(width - width % 4, height, ColorFormatBGR24, img.DataPtrUint8(), faceInfo, EnableAge | EnableGender)
		for idx := 0; idx < int(faceInfo.FaceNum); idx ++{
			rect := image.Rect(int(faceInfo.FaceRect[idx].Left),
				int(faceInfo.FaceRect[idx].Top),
				int(faceInfo.FaceRect[idx].Right),
				int(faceInfo.FaceRect[idx].Bottom))
			// 把人脸框起来
			gocv.Rectangle(img, rect, color.RGBA{G: 255}, 2)
			if err == nil {
				age, _ := engine.GetAge()
				gender, _ := engine.GetGender()
				var ageResult string
				var genderResult string
				if age.AgeArray[idx] <= 0 {
					ageResult = "N/A"
				} else {
					ageResult = strconv.Itoa(int(age.AgeArray[idx]))
				}
				if gender.GenderArray[idx] < 0 {
					genderResult = "N/A"
				} else if gender.GenderArray[idx] == 0 {
					genderResult = "Male"
				} else {
					genderResult = "Female"
				}
				// 把年龄和性别信息绘在图上（!!注意，opencv不支持ASCII以外的字体，如有需要，引入freetype，加载本地字体资源）
				gocv.PutText(img,
					fmt.Sprintf("Age: %s", ageResult),
					image.Pt(int(faceInfo.FaceRect[idx].Right + 2), int(faceInfo.FaceRect[idx].Top + 10)),
					gocv.FontHersheyPlain,
					1,
					color.RGBA{R:255},
					1)
				gocv.PutText(img,
					fmt.Sprintf("Gender: %s", genderResult),
					image.Pt(int(faceInfo.FaceRect[idx].Right + 2), int(faceInfo.FaceRect[idx].Top + 25)),
					gocv.FontHersheyPlain,
					1,
					color.RGBA{R:255},
					1)
			}
		}
	}
}
