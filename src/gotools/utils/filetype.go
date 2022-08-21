package utils

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var fileTypeMap sync.Map

func init() {
	fileTypeMap.Store("ffd8ffe000104a464946", "jpg")  //JPEG (jpg)
	fileTypeMap.Store("89504e470d0a1a0a0000", "png")  //PNG (png)
	fileTypeMap.Store("47494638396126026f01", "gif")  //GIF (gif)
	fileTypeMap.Store("49492a00227105008037", "tif")  //TIFF (tif)
	fileTypeMap.Store("424d228c010000000000", "bmp")  //16色位图(bmp)
	fileTypeMap.Store("424d8240090000000000", "bmp")  //24位位图(bmp)
	fileTypeMap.Store("424d8e1b030000000000", "bmp")  //256色位图(bmp)
	fileTypeMap.Store("41433130313500000000", "dwg")  //CAD (dwg)
	fileTypeMap.Store("3c21444f435459504520", "html") //HTML (html)   3c68746d6c3e0  3c68746d6c3e0
	fileTypeMap.Store("3c68746d6c3e0", "html")        //HTML (html)   3c68746d6c3e0  3c68746d6c3e0
	fileTypeMap.Store("68656c6c6f3c21444f43", "html")
	fileTypeMap.Store("3c21646f637479706520", "htm") //HTM (htm)
	fileTypeMap.Store("48544d4c207b0d0a0942", "css") //css
	fileTypeMap.Store("efbbbf22e6a08", "csv")        //csv
	fileTypeMap.Store("696b2e71623d696b2e71", "js")  //js
	fileTypeMap.Store("7b5c727466315c616e73", "rtf") //Rich Text Format (rtf)
	fileTypeMap.Store("38425053000100000000", "psd") //Photoshop (psd)
	fileTypeMap.Store("46726f6d3a203d3f6762", "eml") //Email [Outlook Express 6] (eml)
	fileTypeMap.Store("d0cf11e0a1b11ae10000", "doc") //MS Excel 注意：word、msi 和 excel的文件头一样
	fileTypeMap.Store("d0cf11e0a1b11ae10000", "vsd") //Visio 绘图
	fileTypeMap.Store("5374616E64617264204A", "mdb") //MS Access (mdb)
	fileTypeMap.Store("252150532D41646F6265", "ps")
	fileTypeMap.Store("255044462d312e350d0a", "pdf")  //Adobe Acrobat (pdf)
	fileTypeMap.Store("2e524d46000000120001", "rmvb") //rmvb/rm相同
	fileTypeMap.Store("464c5601050000000900", "flv")  //flv与f4v相同
	fileTypeMap.Store("00000020667479706d70", "mp4")
	fileTypeMap.Store("49443303000000002176", "mp3")
	fileTypeMap.Store("000001ba210001000180", "mpg") //
	fileTypeMap.Store("3026b2758e66cf11a6d9", "wmv") //wmv与asf相同
	fileTypeMap.Store("52494646e27807005741", "wav") //Wave (wav)
	fileTypeMap.Store("52494646d07d60074156", "avi")
	fileTypeMap.Store("4d546864000000060001", "mid") //MIDI (mid)
	fileTypeMap.Store("504b0304140000000800", "zip")
	fileTypeMap.Store("526172211a070", "rar")
	fileTypeMap.Store("235468697320636f6e66", "ini")
	fileTypeMap.Store("504b03040a0000000000", "jar")
	fileTypeMap.Store("4d5a9000030000000400", "exe")        //可执行文件
	fileTypeMap.Store("3c25402070616765206c", "jsp")        //jsp文件
	fileTypeMap.Store("4d616e69666573742d56", "mf")         //MF文件
	fileTypeMap.Store("3c3f786d6c2076657273", "xml")        //xml文件
	fileTypeMap.Store("494e5345525420494e54", "sql")        //xml文件
	fileTypeMap.Store("7061636b616765207765", "java")       //java文件
	fileTypeMap.Store("7061636b616765207574", "go")         //golang文件
	fileTypeMap.Store("406563686f206f66660d", "bat")        //bat文件
	fileTypeMap.Store("1f8b0800000000000000", "gz")         //gz文件
	fileTypeMap.Store("6c6f67346a2e726f6f74", "properties") //bat文件
	fileTypeMap.Store("cafebabe0000002e0041", "class")      //bat文件
	fileTypeMap.Store("49545346030000006000", "chm")        //bat文件
	fileTypeMap.Store("04000000010000001300", "mxp")        //bat文件
	fileTypeMap.Store("504b0304140006000800", "docx")       //docx文件
	fileTypeMap.Store("d0cf11e0a1b11ae10000", "wps")        //WPS文字wps、表格et、演示dps都是一样的
	fileTypeMap.Store("6431303a637265617465", "torrent")
	fileTypeMap.Store("6D6F6F76", "mov")         //Quicktime (mov)
	fileTypeMap.Store("FF575043", "wpd")         //WordPerfect (wpd)
	fileTypeMap.Store("CFAD12FEC5FD746F", "dbx") //Outlook Express (dbx)
	fileTypeMap.Store("2142444E", "pst")         //Outlook (pst)
	fileTypeMap.Store("AC9EBD8F", "qdf")         //Quicken (qdf)
	fileTypeMap.Store("E3828596", "pwl")         //Windows Password (pwl)
	fileTypeMap.Store("2E7261FD", "ram")         //Real Audio (ram)
}

// 获取前面结果字节的二进制
func bytesToHexString(src []byte) string {
	res := bytes.Buffer{}
	if src == nil || len(src) <= 0 {
		return ""
	}
	temp := make([]byte, 0)
	for _, v := range src {
		sub := v & 0xFF
		hv := hex.EncodeToString(append(temp, sub))
		if len(hv) < 2 {
			res.WriteString(strconv.FormatInt(int64(0), 10))
		}
		res.WriteString(hv)
	}
	return res.String()
}

// 用文件前面几个字节来判断
// fSrc: 文件字节流（就用前面几个字节）
func GetFileTypeCustom(fSrc []byte) string {
	var fileType string
	fileCode := bytesToHexString(fSrc)

	fileTypeMap.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(string)
		if strings.HasPrefix(fileCode, strings.ToLower(k)) ||
			strings.HasPrefix(k, strings.ToLower(fileCode)) {
			fileType = v
			return false
		}
		return true
	})
	// contentType := http.DetectContentType(fSrc[:10]) 返回文件类型

	return fileType
}

func GetFileTypeUseHttp(fSrc []byte) string {
	return http.DetectContentType(fSrc[:10])
}

// example:
// fSrc, _ = ioutil.ReadFile("../../../yipeng888/simhei.ttf")
// 	t.Log(GetFileTypeCustom(fSrc[:10]))
// 	t.Log(GetFileTypeUseHttp(fSrc[:10]))

// 1 JPEG (jpg),文件头:FFD8FF
// 2 PNG (png),文件头:89504E47
// 3 GIF (gif),文件头:47494638
// 4 TIFF (tif),文件头:49492A00
// 5 Windows Bitmap (bmp),文件头:424D
// 6 CAD (dwg),文件头:41433130
// 7 Adobe Photoshop (psd),文件头:38425053
// 8 Rich Text Format (rtf),文件头:7B5C727466
// 9 XML (xml),文件头:3C3F786D6C
// 10 HTML (html),文件头:68746D6C3E
// 11 Email [thorough only] (eml),文件头:44656C69766572792D646174653A
// 12 Outlook Express (dbx),文件头:CFAD12FEC5FD746F
// 13 Outlook (pst),文件头:2142444E
// 14 MS Word/Excel (xls.or.doc),文件头:D0CF11E0
// 15 MS Access (mdb),文件头:5374616E64617264204A
// 16 WordPerfect (wpd),文件头:FF575043
// 17 Postscript (eps.or.ps),文件头:252150532D41646F6265
// 18 Adobe Acrobat (pdf),文件头:255044462D312E
// 19 Quicken (qdf),文件头:AC9EBD8F
// 20 Windows Password (pwl),文件头:E3828596
// 21 ZIP Archive (zip),文件头:504B0304
// 22 RAR Archive (rar),文件头:52617221
// 23 Wave (wav),文件头:57415645
// 24 AVI (avi),文件头:41564920
// 25 Real Audio (ram),文件头:2E7261FD
// 26 Real Media (rm),文件头:2E524D46
// 27 MPEG (mpg),文件头:000001BA
// 28 MPEG (mpg),文件头:000001B3
// 29 Quicktime (mov),文件头:6D6F6F76
// 30 Windows Media (asf),文件头:3026B2758E66CF11
// 31 MIDI (mid),文件头:4D546864
