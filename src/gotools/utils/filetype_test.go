package utils

import (
	"io/ioutil"
	"testing"
)

func TestGetFileType(t *testing.T) {
	//f, err := os.Open("C:\\Users\\Administrator\\Desktop\\api.html")

	fSrc, _ := ioutil.ReadFile("300G.xlsx")
	t.Log(GetFileTypeCustom(fSrc[:10]))
	t.Log(GetFileTypeUseHttp(fSrc[:10]))

	// f, _ := os.Open("./yp.png")
	fSrc, _ = ioutil.ReadFile("filetype.go")
	t.Log(GetFileTypeCustom(fSrc[:10]))
	t.Log(GetFileTypeUseHttp(fSrc[:10]))

	// f, _ = os.Open("./yp.rar")
	fSrc, _ = ioutil.ReadFile("zabbix.doc")
	t.Log(GetFileTypeCustom(fSrc[:10]))
	t.Log(GetFileTypeUseHttp(fSrc[:10]))

	fSrc, _ = ioutil.ReadFile("index.html")
	t.Log(GetFileTypeCustom(fSrc[:10]))
	t.Log(GetFileTypeUseHttp(fSrc[:10]))

	fSrc, _ = ioutil.ReadFile("../../../yipeng888/simhei.ttf")
	t.Log(GetFileTypeCustom(fSrc[:10]))
	t.Log(GetFileTypeUseHttp(fSrc[:10]))
	// defer fSrc.Close()

}
