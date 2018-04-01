package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/syncmap"

	"github.com/PuerkitoBio/goquery"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/phayes/freeport"
	"github.com/vjeantet/chromedp"
	"github.com/vjeantet/chromedp/runner"
)

// var cm *chromedp.CDP
// var cs.ctx context.Context

type ChromeShot struct {
	cdp    *chromedp.CDP
	ctx    context.Context
	cancel context.CancelFunc
	Logf   func(string, ...interface{})
}

func NewChromeShot(showBrowser bool, logf func(string, ...interface{})) (*ChromeShot, error) {
	// var
	// create context
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// create chrome instance
	// cm, err = chromedp.New(cs.ctx, chromedp.WithLogf(l.Printf))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Start Webserver

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(os.Getenv("BF_COMMONS_PATH")+string(os.PathSeparator)+"public"))))
	http.HandleFunc("/", helloWorld)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	baseURL = "http://" + listener.Addr().String()
	logf("baseURL=%s", baseURL)

	go http.Serve(listener, nil)

	port, err := freeport.GetFreePort()
	if err != nil {
		logf(err.Error())
	}

	cdp, err := chromedp.New(ctx, chromedp.WithRunnerOptions(
		runner.Flag("headless", !showBrowser),
		runner.Flag("disable-gpu", true),
		runner.Flag("no-first-run", true),
		runner.Flag("no-default-browser-check", true),
		runner.Port(port),
	), chromedp.WithLog(log.New(ioutil.Discard, "", 0).Printf))
	// ), chromedp.WithLog(log.New(os.Stderr, "", 0).Printf))
	if err != nil {
		return nil, err
	}
	return &ChromeShot{
		cdp:    cdp,
		ctx:    ctx,
		cancel: cancel,
		Logf:   logf,
	}, nil
}

var htmlMap = syncmap.Map{}
var baseURL = ""

func helloWorld(w http.ResponseWriter, r *http.Request) {
	hUrl := strings.ToLower(r.URL.Path)
	hUrl = strings.TrimLeft(hUrl, "/")

	if htmlContent, ok := htmlMap.Load(hUrl); ok {
		w.Write(htmlContent.([]byte))
	} else {
		// cs.Logf("not found - ", hUrl)
		w.WriteHeader(404)
		w.Write([]byte("Not Found !!!"))
	}
}

func (c *ChromeShot) revokeUrl(uid string) {
	htmlMap.Delete(uid)
}

func (c *ChromeShot) urlForHtmlContent(content string) (string, string, error) {
	// Set URL
	uid, err := uuid.NewV4()
	if err != nil {
		return "", "", err
	}

	// Set Content for URL
	htmlMap.Store(uid.String(), []byte(content))

	return baseURL + "/" + uid.String(), uid.String(), nil
}

func (cs *ChromeShot) EmbedImageForDomElements(htmlContent string, selectors []string) (string, error) {
	urlStr, uid, err := cs.urlForHtmlContent(htmlContent)
	if err != nil {
		return "", err
	}

	cs.Logf("urlStr--> %s", urlStr)

	// Check ID presence
	doc, err := goquery.NewDocument(urlStr)
	if err != nil {
		return "", err
	}
	idsOk := []string{}
	for i, id := range selectors {
		sel := doc.Find(id)

		if len(sel.Nodes) != 1 {
			cs.Logf("Unique HTML element [%s] not found in document", id)
		} else {
			idsOk = append(idsOk, selectors[i])
		}
	}

	cs.cdp.Run(cs.ctx, chromedp.Navigate(urlStr))
	cs.cdp.Run(cs.ctx, chromedp.Sleep(2*time.Second))
	datas := map[string][]byte{}
	cs.Logf("selectors -->%s", idsOk)
	for k, id := range idsOk {
		cs.Logf("k,selector-->%d,%s", k, id)
		var buf []byte

		cctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		if err := cs.cdp.Run(cctx, chromedp.WaitVisible(id, chromedp.BySearch)); err != nil {
			cs.Logf("err-->", err)
			continue
		}

		cs.cdp.Run(cs.ctx, chromedp.Screenshot(id, &buf, chromedp.NodeVisible, chromedp.BySearch))
		// cs.cdp.Run(cs.ctx, chromedp.CaptureScreenshot(&buf))
		// cs.Logf(string(buf))
		datas[id] = buf
	}
	cs.Logf("--> END")

	// doc, err := goquery.NewDocument(urlStr)
	if err != nil {
		return "", err
	}

	for i, v := range datas {
		// Write To Disk
		//ioutil.WriteFile(slugify.Slugify(i)+".png", v, 0644)
		// cs.Logf("i=", i)
		// Replace HTML
		buf64 := base64.StdEncoding.EncodeToString(v)
		sel := doc.Find(i)
		for k := range sel.Nodes {
			single := sel.Eq(k)
			single.ReplaceWithHtml(`<img src="data:image/png;base64,` + buf64 + `" />`)
		}
	}
	// cs.Logf("final")

	// remove url content form memory
	cs.revokeUrl(uid)

	return doc.Html()
}

func (cs *ChromeShot) Stop() error {
	cs.cdp.Kill()
	return nil
	// cs.cancel()
	// // shutdown chrome
	// err := cs.cdp.Shutdown(cs.ctx)
	// if err != nil {
	// 	return err
	// }

	// don't wait for chrome to finish
	// cs.cdp.Kill()
	// return nil
}
