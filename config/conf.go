package config

import (
	"context"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Input struct {
	DUrl       string //单个输入URL
	UrlsFile   string //输入urls.txt
	CookieFile string //输入cookie.txt
	Seq        string //页面范围 4:434
	SeqStart   int    //页面范围 4:434
	SeqEnd     int    //页面范围 4:434
	Volume     int    //多册，只下第N册
	Speed      uint   //限速
	SaveFolder string //下载文件存放目录，默认为当前文件夹下 Downloads 目录下
	//;生成 dezoomify-rs 可用的文件(默认生成文件名 dezoomify-rs.urls.txt）
	// ;0 = 禁用，1=启用 （只对支持的图书馆有效）
	Format        string //;全高清图下载时，指定宽度像素（16开纸185mm*260mm，像素2185*3071）
	UserAgent     string //自定义UserAgent
	AutoDetect    int    //自动检测下载URL。可选值[0|1|2]，;0=默认;1=通用批量下载（类似IDM、迅雷）;2= IIIF manifest.json 自动检测下载图片
	MergePDFs     bool   //;台北故宫博物院 - 善本古籍，是否整册合并一个PDF下载？0=否，1=是。整册合并一个PDF遇到某一册最后一章节【无影像】会导致下载失败。 如：新刊校定集注杜詩 三十六卷 第二十四冊 聞惠子過東溪 无影像
	DezoomifyPath string //dezoomify-rs 本地目录位置
	DezoomifyRs   string //dezoomify-rs 参数
	UseDziRs      bool   //启用DezoomifyRs下载IIIF
	FileExt       string //指定下载的扩展名
	Threads       uint
	Timeout       float64
	Retry         int  //重试次数
	Bookmark      bool //只下載書簽目錄（浙江寧波天一閣）
	Help          bool
	Version       bool
}

func Init(ctx context.Context) bool {

	dir, _ := os.Getwd()

	//你们为什么没有良好的电脑使用习惯？中文虽好，但不适用于计算机。
	if os.PathSeparator == '\\' {
		matched, _ := regexp.MatchString(`([^A-z0-9_\\/\-:.]+)`, dir)
		if matched {
			fmt.Println("本软件存放目录，不能包含空格、中文等特殊符号。推荐：D:\\bookget")
			fmt.Println("按回车键终止程序。Press Enter to exit ...")
			endKey := make([]byte, 1)
			os.Stdin.Read(endKey)
			os.Exit(0)
		}
	}
	iniConf, _ := initINI()

	flag.StringVar(&Conf.UrlsFile, "i", iniConf.UrlsFile, "下载的URLs，指定任意本地文件，例如：urls.txt")
	flag.StringVar(&Conf.SaveFolder, "o", iniConf.SaveFolder, "下载保存到目录")
	flag.StringVar(&Conf.Seq, "seq", iniConf.Seq, "页面范围，如4:434")
	flag.IntVar(&Conf.Volume, "vol", iniConf.Volume, "多册图书，只下第N册")
	flag.StringVar(&Conf.Format, "fmt", iniConf.Format, "IIIF 图像请求URI: full/full/0/default.jpg")
	flag.StringVar(&Conf.UserAgent, "ua", iniConf.UserAgent, "user-agent")
	flag.BoolVar(&Conf.MergePDFs, "mp", iniConf.MergePDFs, "合并PDF文件下载，可选值[0|1]。0=否，1=是。仅对 rbk-doc.npm.edu.tw 有效。")
	flag.BoolVar(&Conf.Bookmark, "mark", iniConf.Bookmark, "只下载书签目录，可选值[0|1]。0=否，1=是。仅对 gj.tianyige.com.cn 有效。")
	flag.BoolVar(&Conf.UseDziRs, "dzi", iniConf.UseDziRs, "使用dezoomify-rs下载，仅对支持iiif的网站生效。")
	flag.StringVar(&Conf.CookieFile, "c", iniConf.CookieFile, "指定cookie.txt文件路径")
	flag.StringVar(&Conf.FileExt, "ext", iniConf.FileExt, "指定文件扩展名[.jpg|.tif|.png]等")
	flag.UintVar(&Conf.Threads, "n", iniConf.Threads, "最大并发连接数")
	flag.Float64Var(&Conf.Timeout, "timeout", iniConf.Timeout, "最大并发连接数")
	flag.UintVar(&Conf.Speed, "speed", iniConf.Speed, "下载限速 N 秒/任务，cuhk推荐5-60")
	flag.IntVar(&Conf.Retry, "r", iniConf.Retry, "下载重试次数")
	flag.IntVar(&Conf.AutoDetect, "a", iniConf.AutoDetect, "自动检测下载URL。可选值[0|1|2]，;0=默认;\n1=通用批量下载（类似IDM、迅雷）;\n2= IIIF manifest.json 自动检测下载图片")
	flag.BoolVar(&Conf.Help, "h", false, "显示帮助")
	flag.BoolVar(&Conf.Version, "v", false, "显示版本")
	flag.StringVar(&Conf.DezoomifyRs, "rs", iniConf.DezoomifyRs, "dezoomify-rs 参数")
	Conf.DezoomifyPath = iniConf.DezoomifyPath
	flag.Parse()

	k := len(os.Args)
	if k == 2 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			printVersion()
			return false
		}
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			printHelp()
			return false
		}
	}
	v := flag.Arg(0)
	if strings.HasPrefix(v, "http") {
		Conf.DUrl = v
	}
	if Conf.UrlsFile != "" && !strings.Contains(Conf.UrlsFile, string(os.PathSeparator)) {
		Conf.UrlsFile = dir + string(os.PathSeparator) + Conf.UrlsFile
	}
	//fmt.Printf("%+v", Conf)
	if Conf.Speed > 60 {
		Conf.Speed = 60
	}
	initSeq()
	//保存目录处理
	_ = os.Mkdir(Conf.SaveFolder, os.ModePerm)
	return true
}

func initINI() (io Input, err error) {
	dir, _ := os.Getwd()
	fPath, _ := os.Executable()
	binDir := filepath.Dir(fPath)
	var iniFile string
	fi, err := os.Stat("/etc/bookget/config.ini")
	if string(os.PathSeparator) == "/" && err == nil && fi.Size() > 0 {
		iniFile = "/etc/bookget/config.ini"
	} else {
		iniFile = binDir + string(os.PathSeparator) + "config.ini"
	}
	cFile := dir + string(os.PathSeparator) + "cookie.txt"
	urls := dir + string(os.PathSeparator) + "urls.txt"
	c := uint(runtime.NumCPU() * 2)

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/118.0"
	format := "full/full/0/default.jpg"
	io = Input{
		DUrl:          "",
		UrlsFile:      urls,
		CookieFile:    cFile,
		Seq:           "",
		SeqStart:      0,
		SeqEnd:        0,
		Volume:        0,
		Speed:         0,
		SaveFolder:    dir,
		Format:        format,
		UserAgent:     ua,
		AutoDetect:    0,
		MergePDFs:     true,
		DezoomifyPath: "",
		DezoomifyRs:   "-l --compression 20",
		UseDziRs:      false,
		FileExt:       ".jpg",
		Threads:       c,
		Timeout:       3000,
		Retry:         3,
		Bookmark:      false,
		Help:          false,
		Version:       false,
	}

	if os.PathSeparator == '\\' {
		io.DezoomifyPath = "dezoomify-rs.exe"
		if fi, err := os.Stat(dir + "\\dezoomify-rs.exe"); err == nil && fi.Size() > 0 {
			io.DezoomifyPath = dir + "\\dezoomify-rs.exe"
		}
	} else {
		io.DezoomifyPath = "dezoomify-rs"
		if fi, err := os.Stat(dir + "/dezoomify-rs"); err == nil && fi.Size() > 0 {
			io.DezoomifyPath = dir + "/dezoomify-rs"
		}
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, iniFile)
	if err != nil {
		return
	}

	io.AutoDetect = cfg.Section("").Key("app_mode").MustInt(0)
	io.SaveFolder = cfg.Section("paths").Key("data").String()
	if io.SaveFolder == "" {
		io.SaveFolder = dir
	}

	secDown := cfg.Section("download")
	io.FileExt = secDown.Key("ext").String()
	io.Threads = secDown.Key("threads").MustUint(c)
	if io.Threads == 0 {
		io.Threads = c
	}

	io.Timeout = secDown.Key("timeout").MustFloat64()
	// 超时不可以短于10秒
	if io.Timeout < 10 {
		io.Timeout = 3000
	}

	io.Speed = secDown.Key("speed").MustUint(c)

	secCus := cfg.Section("custom")
	io.Seq = secCus.Key("seq").String()
	io.Volume = secCus.Key("vol").MustInt(0)
	io.MergePDFs = secCus.Key("mp").MustBool(true)
	io.Bookmark = secCus.Key("bookmark").MustBool(false)
	io.UserAgent = secCus.Key("ua").MustString(ua)

	secDzi := cfg.Section("dzi")
	io.UseDziRs = secDzi.Key("dzi").MustBool(false)
	io.DezoomifyRs = secDzi.Key("rs").String()
	io.Format = secDzi.Key("format").MustString(format)

	if !strings.Contains(io.DezoomifyRs, "-n") && !strings.Contains(io.DezoomifyRs, "--parallelism") {
		io.DezoomifyRs += fmt.Sprintf(" -n=%d", io.Threads)
	}
	return io, nil
}

func printHelp() {
	printVersion()
	fmt.Println(`Usage: bookget [OPTION]... [URL]...`)
	flag.PrintDefaults()
	fmt.Println("Email bug reports, questions, discussions to zhudw<zhudwi@foxmail.com>")
	fmt.Println("and/or open issues at https://github.com/deweizhu/bookget/issues")
}

func printVersion() {
	fmt.Printf("bookget v%s\n", version)
}
