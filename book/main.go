package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/extensions"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type Category struct {
	Title	string
	Code	string
	Cid     string
	Pid  	string
}

var proxies []*url.URL = []*url.URL{
	&url.URL{Host: "180.109.241.71:4236"},
	&url.URL{Host: "60.177.88.126:4240"},
	&url.URL{Host: "106.59.59.30:4237"},
	&url.URL{Host: "123.179.128.158:4206"},
	&url.URL{Host: "123.179.128.16:4206"},
}

func randomProxySwitcher(_ *http.Request) (*url.URL, error) {
	return proxies[rand.Intn(len(proxies))], nil
}

func main()  {

	Db, errs := sql.Open("mysql", "root:root@tcp(127.0.0.1)/category")
	if errs != nil {
		fmt.Printf("%s",errs.Error())
	}
	defer Db.Close()


	var url = make(map[string]bool)
	var data []Category

	c := colly.NewCollector(
		//colly.MaxDepth(5),
		//colly.Async(true),
		colly.Debugger(&debug.LogDebugger{}),
	)

	/*rp, err := proxy.RoundRobinProxySwitcher(
		"socks5://106.56.90.58:4256",
		"socks5://115.212.36.155:4203",
		"socks5://114.230.68.35:4228",
		"socks5://125.125.67.241:4208",
		"socks5://119.120.248.24:4214",
		)
	if err != nil {
		log.Fatal(err)
	}*/


	// ...
	c.SetProxyFunc(randomProxySwitcher)
	//c.SetProxyFunc(rp)
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})
	extensions.RandomUserAgent(c)
	extensions.Referer(c)


	c.OnHTML(".category-list .category-item a",func(e *colly.HTMLElement) {

		href := e.Attr("href")

		code := e.ChildText(".category-code")
		title := e.ChildText(".category-title")

		re, err := regexp.Compile(`\d+`)

		if err != nil{
			fmt.Printf("%s\n",err.Error())
			return
		}
		//查找符合正则的第一个
		all := re.FindAll([]byte(href),-1)
		cid := string(all[0])
		href += "?pid="+cid
		//href = "http://ztflh.xhma.com"+href
		//var mutex sync.Mutex
		//mutex.Lock()

		pid  := e.Request.URL.Query().Get("pid")
		if pid == ""{
			pid = "1"
		}
		//pid := query["pid"][0]
		//fmt.Printf("%s,%s\n",pid,cid)



		data = append(data,Category{title,code,cid,pid})
		//fmt.Printf("%s,%s,%s\n",href,host,strings.Contains(href,host))



		if url[cid] == false {
			e.Request.Visit(href)
			url[cid] = true
		}
		//mutex.Unlock()
	})

	c.OnError(func(rp *colly.Response, err error) {
		c.Visit(rp.Request.URL.String())
		fmt.Println("Something went wrong:", err.Error())
	})

	c.Visit("http://ztflh.xhma.com/")
	//c.Wait()


	fmt.Printf("%s\n",len(data))

	// 3185 2547

	stmt,  err := Db.Prepare(`INSERT INTO category (name, code, cid,pid) VALUES (?,?, ?, ?)`)
	if err != nil {
		fmt.Printf("%s",err.Error())
		return
	}

	defer stmt.Close()

	for _,v := range data{
		_, err = stmt.Exec(v.Title, v.Code,v.Cid,v.Pid)
		if err != nil {
			fmt.Printf("%s", err.Error())
			return
		}
	}


}
