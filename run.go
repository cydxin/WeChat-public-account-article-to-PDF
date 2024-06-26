package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"time"
)

var title string

func main() {
	listByte, err := os.ReadFile("./url.txt")
	if err != nil {
		return
	}
	list := strings.Split(string(listByte), "\n")
	fmt.Println("共", len(list), "个")
	//var wg sync.WaitGroup
	for i, s := range list {
		//wg.Add(1) The filename, directory name, or volume label syntax is incorrect.
		//go func(s string) {
		// create context
		url := strings.TrimPrefix(s, "")
		ctx, cancel := chromedp.NewContext(context.Background())
		ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
		// capture pdf
		fmt.Println("第", i, "个")
		var buf []byte
		if err := chromedp.Run(ctx, printToPDF(
			url,
			&buf, &title)); err != nil {
			log.Fatal(err)
		}
		title = strings.ReplaceAll(title, "/", "-")
		title = strings.ReplaceAll(title, ":", "：")
		title = strings.ReplaceAll(title, "*", "-")
		title = strings.ReplaceAll(title, "?", "？")
		title = strings.ReplaceAll(title, `"`, "'")
		title = strings.ReplaceAll(title, "<", "《")
		title = strings.ReplaceAll(title, ">", "》")
		if err := os.WriteFile(title+".pdf", buf, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println("写入 ", title, ".pdf")
		ctx.Done()
		cancel()
		cancelTimeout()
		//wg.Done()
		//}(s)
	}
	//wg.Wait()
	fmt.Println("宝，结束辣")
	time.Sleep(50 * time.Second)
}

// slowScrollToBottom 缓慢滚动到页面底部的操作
func slowScrollToBottomctx(ctx context.Context) {
	jsStr := `var i = 2
    var element = document.documentElement
    element.scrollTop = 0;  // 不管他在哪里，都让他先回到最上面
 
    // 设置定时器，时间即为滚动速度
    function main() {
        if (element.scrollTop + element.clientHeight == element.scrollHeight) {
            clearInterval(interval)
            console.log('已经到底部了')
        } else {
            // 300 代表每次移动300px
            element.scrollTop += 300;
            console.log(i);
            i += 1;
        }
    }
    // 定义ID 200代表300毫秒滚动一次
    interval = setInterval(main, 100)`

	lengthStr := `var element = document.documentElement
	element.scrollTop = 0;
	element.scrollHeight
	`

	lengthInt := 0
	err := chromedp.Run(ctx,
		chromedp.Evaluate(lengthStr, &lengthInt),
	)
	if err != nil {
		log.Fatal("123:", err)
	}

	//循环滚轮实现
	if lengthInt > 0 {
		fmt.Println("页高度：", lengthInt)
		err = chromedp.Run(ctx,
			chromedp.Evaluate(jsStr, nil),
		)

		sleepDuration := time.Duration(lengthInt/300*100) * time.Millisecond
		fmt.Println("翻页时间：", sleepDuration)
		time.Sleep(sleepDuration)
		//无效代码 纯闲得 看看滚动条位置
		topInt := 0
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`element.scrollTop`, &topInt),
		)
	}

}

// printToPDF 打印特定的 PDF 页面
func printToPDF(urlstr string, res *[]byte, title *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			//获取长度
			slowScrollToBottomctx(ctx)
			return nil
		}),
		chromedp.Text(`#activity-name`, title, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			*res, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	}
}
