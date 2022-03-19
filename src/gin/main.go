package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/ffhelicopter/tmm/api"
	"github.com/ffhelicopter/tmm/handler"
	"github.com/gin-gonic/gin"
)

func submit(c *gin.Context) {
	name := c.DefaultQuery("name", "lily")
	c.String(200, fmt.Sprintf("hello %s\n", name))
}

// 定义全局的CORS中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}
func LimitHandler(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			c.Data(httpError.StatusCode, lmt.GetMessageContentType(), []byte(httpError.Message))
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func upload(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get err %s", err.Error()))
	}
	// 获取所有图片
	files := form.File["files"]
	// 遍历所有图片
	for _, file := range files {
		// 逐个存
		// if err := c.SaveUploadedFile(file, file.Filename); err != nil {
		dst := fmt.Sprintf("./website/tmp/%s", file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload err %s", err.Error()))
			return
		}
	}
	c.String(200, fmt.Sprintf("upload ok %d files", len(files)))

}

func ginInit() {
	router := gin.Default()
	// 静态资源加载，本例为css,js以及资源图片
	router.StaticFS("/public", http.Dir("./website/static"))
	router.StaticFile("/favicon.jpg", "./website/favicon.jpg")
	router.StaticFile("/uploads", "./website/tpl/upload.html")

	// 导入所有模板，多级目录结构需要这样写,tpl动态目录
	router.LoadHTMLGlob("./website/tpl/*")

	// 使用全局CORS中间件。
	// router.Use(Cors())
	// 即使是全局中间件，在use前的代码不受影响
	//rate-limit 限流中间件
	lmt := tollbooth.NewLimiter(1, nil)
	lmt.SetMessage("服务繁忙，请稍后再试...")

	// website分组
	v := router.Group("/")
	{
		v.GET("/index.html", LimitHandler(lmt), handler.IndexHandler)
		v.GET("/add.html", handler.AddHandler)
		v.POST("/postme.html", handler.PostmeHandler)
		v.POST("/submit", submit)
		v.POST("/upload", upload)
	}

	v1 := router.Group("/v1")
	{
		// 下面是群组中间的用法
		// v1.Use(Cors())
		// 单个中间件的用法
		// v1.GET("/user/:id/*action",Cors(), api.GetUser)
		// rate-limit
		v1.GET("/user/:id/*action", LimitHandler(lmt), api.GetUser)
		//v1.GET("/user/:id/*action", Cors(), api.GetUser)
		// AJAX OPTIONS ，下面是有关OPTIONS用法的示例
		// v1.OPTIONS("/users", OptionsUser)      // POST
		// v1.OPTIONS("/users/:id", OptionsUser)  // PUT, DELETE
	}

	// router.Run(":80")
	// 这样写就可以了，下面所有代码（go1.8+）是为了优雅处理重启等动作。
	srv := &http.Server{
		Addr:         ":80",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		// 监听请求
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 优雅Shutdown（或重启）服务
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt) // syscall.SIGKILL
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		fmt.Println("timeout")
	}
	log.Println("Server exiting")

	// Listen and serve on 0.0.0.0:8000
	// router.Run(":8000")
}

// var router *gin.Engine

func sortInit() {
	listInt := []int{5, 9, 1, 63, 8, 14, 6, 49, 25, 4, 36, 3, 73, 0, 16}
	list := listInt
	// s := sorts.SortInt{}
	// h := NewHeap(listInt)
	// sortS := s.BubblingSort(listInt)
	// fmt.Println("sortS:", sortS)
	// ss := s.SelectSort(listInt)
	// fmt.Println("select:", ss)
	// fmt.Println("InsertSort:", s.InsertSort(listInt))
	// fmt.Println("ShellSort:", s.ShellSort(listInt))
	// for _, v := range list {
	// 	h.Push(v)
	// 	fmt.Println("1:,", list)
	// }

	// // 将堆元素移除
	// for range list {
	// 	h.Pop()
	// }
	heapSort(list)
	fmt.Println("heapSort:", list)
}

func main() {

	// ginInit()

	sortInit()
	fmt.Println("m/2", 12/8)

}

// 一个最大堆，一颗完全二叉树
// 最大堆要求节点元素都不小于其左右孩子
type Heap struct {
	// 堆的大小
	Size int
	// 使用内部的数组来模拟树
	// 一个节点下标为 i，那么父亲节点的下标为 (i-1)/2
	// 一个节点下标为 i，那么左儿子的下标为 2i+1，右儿子下标为 2i+2
	Array []int
}

// 初始化一个堆
func NewHeap(array []int) *Heap {
	h := new(Heap)
	h.Array = array
	return h
}

// 最大堆插入元素
func (h *Heap) Push(x int) {
	// 堆没有元素时，使元素成为顶点后退出
	if h.Size == 0 {
		h.Array[0] = x
		h.Size++
		return
	}

	// i 是要插入节点的下标
	i := h.Size

	// 如果下标存在
	// 将小的值 x 一直上浮
	for i > 0 {
		// parent为该元素父亲节点的下标
		parent := (i - 1) / 2

		// 如果插入的值小于等于父亲节点，那么可以直接退出循环，因为父亲仍然是最大的
		if x <= h.Array[parent] {
			break
		}

		// 否则将父亲节点与该节点互换，然后向上翻转，将最大的元素一直往上推
		h.Array[i] = h.Array[parent]
		i = parent
	}

	// 将该值 x 放在不会再翻转的位置
	h.Array[i] = x

	// 堆数量加一
	h.Size++
}

// 最大堆移除根节点元素，也就是最大的元素
func (h *Heap) Pop() int {
	// 没有元素，返回-1
	if h.Size == 0 {
		return -1
	}

	// 取出根节点
	ret := h.Array[0]

	// 因为根节点要被删除了，将最后一个节点放到根节点的位置上
	h.Size--
	x := h.Array[h.Size]  // 将最后一个元素的值先拿出来
	h.Array[h.Size] = ret // 将移除的元素放在最后一个元素的位置上

	// 对根节点进行向下翻转，小的值 x 一直下沉，维持最大堆的特征
	i := 0
	for {
		// a，b为下标 i 左右两个子节点的下标
		a := 2*i + 1
		b := 2*i + 2

		// 左儿子下标超出了，表示没有左子树，那么右子树也没有，直接返回
		if a >= h.Size {
			break
		}

		// 有右子树，拿到两个子节点中较大节点的下标
		if b < h.Size && h.Array[b] > h.Array[a] {
			a = b
		}

		// 父亲节点的值都大于或等于两个儿子较大的那个，不需要向下继续翻转了，返回
		if x >= h.Array[a] {
			break
		}

		// 将较大的儿子与父亲交换，维持这个最大堆的特征
		h.Array[i] = h.Array[a]

		// 继续往下操作
		i = a
	}

	// 将最后一个元素的值 x 放在不会再翻转的位置
	h.Array[i] = x
	return ret
}

func heapSort(arr []int) []int {
	arrLen := len(arr)
	buildMaxHeap(arr, arrLen)
	for i := arrLen - 1; i >= 0; i-- {
		// fmt.Println("arr:", arr)
		swap(arr, 0, i)
		arrLen -= 1
		heapify(arr, 0, arrLen)
	}
	return arr
}

func buildMaxHeap(arr []int, arrLen int) {
	for i := arrLen / 2; i >= 0; i-- {
		heapify(arr, i, arrLen)
	}
}

func heapify(arr []int, i, arrLen int) {
	left := 2*i + 1
	right := 2*i + 2
	largest := i
	if left < arrLen && arr[left] > arr[largest] {
		largest = left
	}
	if right < arrLen && arr[right] > arr[largest] {
		largest = right
	}
	fmt.Println("arr:", arr)
	if largest != i {
		swap(arr, i, largest)
		heapify(arr, largest, arrLen)

	}

}

func swap(arr []int, i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}
