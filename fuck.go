
package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"strconv"
	"strings"

	"crypto/ecdsa"
	crand "crypto/rand"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
)

//错误处理
func handle(why string, e error) {
	if e != nil {
		fmt.Println(why, "错误为：", e)
	}
}

func main() {

	routineChan := make(chan struct{}, 10)
	//如果有余额就将私钥和地址存入文件中
	file, e := os.OpenFile("./addr_amount.txt", os.O_WRONLY|os.O_CREATE, 0761)
	handle("文件打开失败！", e)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	//连接以太坊浏览器查询生成地址余额是否大于0
	bitcoinTypes := []string{"eth", "okc", "bsc", "polygon"}
	link := "https://www.oklink.com/zh-cn/"
	for {
		privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), crand.Reader)
		if err != nil {
			return
		}
		address := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
		addr := address.String()
		fmt.Printf("privateKey: %x", privateKeyECDSA.D.Bytes())
		fmt.Println("\taddr :", addr)
		routineChan <- struct{}{}
		for _, bitcoinType := range bitcoinTypes {
			go func(bitcoinType string) {
				ctx, cancel := chromedp.NewContext(context.Background())
				url := link + bitcoinType + "/address/" + addr
				contentString := ""
				err := chromedp.Run(ctx, chromedp.Navigate(url),
					chromedp.WaitVisible(".align-items-center .color-000000"),
					chromedp.OuterHTML(".align-items-center .color-000000", &contentString),
				)
				if err != nil {
					fmt.Println("failed to run chromdp : ", err)
					return
				}
				fmt.Println("contentstring is : ", contentString)
				defer func() {
					<-routineChan
					cancel()
				}()

				doc, err := goquery.NewDocumentFromReader(strings.NewReader(contentString))
				if err != nil {
					fmt.Println("failed to goquery parse: ", err)
					return
				}
				doc.Find(".color-000000").Each(func(i int, selection *goquery.Selection) {
					text := selection.Text()
					balanceStr := strings.Split(text, " ")
					balance, err := strconv.ParseFloat(balanceStr[0], 10)
					if err != nil {
						fmt.Println("failed to parseFloat ", err, balanceStr)
						return
					}
					if balance > 0 {
						fmt.Printf("oh got it  find the balance is %f\t addr is: %s\n", balance, addr)
						_, err := file.WriteString("pri: " + privateKeyECDSA.D.String() + "addr :" + addr + "\n")
						if err != nil {
							fmt.Println("failed to writeString: ", err)
							return
						}
					}
				})
			}(bitcoinType)
		}
	}
}
