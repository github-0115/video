package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	grpc "video/grpc"
)

var (
	inputPath = flag.String("input", "", "read file path")
	outputDir = flag.String("output", "", "file save dir path")
	threshold = flag.Float64("threshold", 0, "threshold value ")
	addr      = flag.String("addr", "", "client address")
)

func main() {

	flag.Parse()
	if *inputPath == "" {
		fmt.Println("必须指定要文件地址")
		return
	}

	if *outputDir == "" {
		fmt.Println("必须指定文件保存目录")
		return
	}

	if *addr == "" {
		fmt.Println("必须指定client address")
		return
	}

	grpc.InitDedupClient(*addr)

	fileLists := Readlist(*inputPath)
	fmt.Println(len(fileLists))
	framesPaths := make([]string, 0, 0)
	for index, file := range fileLists {
		framesPath, err := DecodeAll(file, *outputDir+"/"+strconv.Itoa(index))
		if err != nil {
			fmt.Println("DecodeAll " + strconv.Itoa(index) + file + " err" + err.Error())
			continue
		}
		framesPaths = append(framesPaths, framesPath)
	}

	for _, framesPath := range framesPaths {
		Decode(framesPath, filepath.Join(framesPath, ".."), *threshold)
	}

}

func DecodeAll(videoPath string, saveDir string) (string, error) {
	//	var path = "C:\\Users\\deepir\\Desktop\\videotest\\test"
	//	return path, nil

	framePath := saveDir + "/frames"
	//判断目录是否存在
	err := CheckSaveDir(framePath)
	if err != nil {
		fmt.Println("check saveDir error " + err.Error())
		return "", err
	}

	cmd := exec.Command("ffmpeg", "-i", videoPath, "-f", "image2", "-vf", "fps=fps=1", filepath.Join(framePath, "%10d.png"))
	// 获取输出对象，可以从该对象中读取输出结果
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	// 保证关闭输出流
	defer stdout.Close()
	// 运行命令
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	log.Println(string(opBytes))
	return "", nil
}

func Decode(path string, saveDir string, threshold float64) error {
	//判断目录是否存在
	err := CheckSaveDir(saveDir)
	if err != nil {
		fmt.Println("check saveDir error " + err.Error())
		return err
	}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("read frames dir error " + err.Error())
		return err
	}

	for idx, info := range infos {
		imagePath, err := filepath.Abs(filepath.Join(path, info.Name()))
		fmt.Println(imagePath)
		if err != nil {
			fmt.Println("read frame path error " + err.Error())
			continue
		}
		frame, err := ReadFrame(imagePath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Read image file %s failed ", imagePath) + err.Error())
			continue
		}
		//
		similar := ContrastFrame(path, idx, frame, threshold)
		if !similar {
			SaveFrame(imagePath, saveDir)
		}
	}

	//delete frames
	return nil
}

func ContrastFrame(videoId string, frameId int, currentFrame []byte, threshold float64) bool {
	_, score, err := grpc.IsStaticVideo(videoId, int64(frameId), currentFrame)
	if err != nil {
		return false
	}
	return score >= threshold
}

func ReadFrame(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func SaveFrame(imagePath string, saveDir string) error {
	_, name := filepath.Split(imagePath)
	if name == "" {
		//TODO return error
	}
	var outPut = filepath.Join(saveDir, name)
	dest, err := os.Create(outPut)
	if err != nil {
		//
		return err
	}
	defer dest.Close()

	image, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer image.Close()

	_, err = io.Copy(dest, image)
	if err != nil {
		return err
	}

	return nil
}

func Readlist(path string) []string {

	fileList := make([]string, 0, 0)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("read file error " + err.Error())
		return fileList
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileList = append(fileList, scanner.Text())
	}

	return fileList
}

func CheckSaveDir(saveDir string) error {

	isExist, err := PathExists(saveDir)
	if err != nil {
		fmt.Println("check saveDir isExist error " + err.Error())
		return err
	}

	if !isExist {
		err := os.MkdirAll(saveDir, 0777)
		if err != nil {
			fmt.Printf("mkdir saveDir error " + err.Error())
			return err
		}
		return nil
	}

	return err
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
