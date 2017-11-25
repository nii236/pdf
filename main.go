package main

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"sync"

	"github.com/wirepair/gcd/gcdapi"

	"github.com/wirepair/gcd"
)

var defaultFlags = []string{
	"--allow-running-insecure-content",
	"--disable-extensions",
	"--disable-gpu",
	"--disable-new-tab-first-run",
	"--disable-notifications",
	"--headless",
	"--ignore-certificate-errors",
	"--no-default-browser-check",
	"--no-first-run",
}

type Client struct {
	*gcd.Gcd
	Tab *gcd.ChromeTarget
}

func main() {
	client := NewClient("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	defer client.ExitProcess()
	client.Render("http://localhost:8081")

}

func NewClient(chromePath string) *Client {
	client := gcd.NewChromeDebugger()
	client.AddFlags(defaultFlags)
	client.StartProcess(chromePath, "/tmp", "9999")
	return &Client{Gcd: client}
}

func (c *Client) Prepare() {
	t, err := c.NewTab()
	if err != nil {
		panic(err)
		panic(err)
	}
	t.CSS.Enable()
	t.DOM.Enable()
	t.Network.Enable(-1, -1)
	t.Page.Enable()
	t.Runtime.Enable()
	c.Tab = t
}

func (c *Client) Render(url string) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	target, err := c.NewTab()
	if err != nil {
		panic(err)

	}
	target.Subscribe("Page.loadEventFired", func(targ *gcd.ChromeTarget, v []byte) {
		options := &gcdapi.PagePrintToPDFParams{
			Landscape:           true,
			DisplayHeaderFooter: false,
			PrintBackground:     true,
			// Scale:               1.0,
			// PaperWidth:              8.5,
			// PaperHeight:             11,
			MarginTop:               0.5,
			MarginBottom:            0.5,
			MarginLeft:              0.5,
			MarginRight:             0.5,
			PageRanges:              "",
			IgnoreInvalidPageRanges: true,
		}
		result, err := targ.Page.PrintToPDFWithParams(options)
		if err != nil {
			panic(err)
		}
		b, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			panic(err)
		}
		binary.Write(os.Stdout, binary.LittleEndian, b)
		wg.Done() // page loaded, we can exit now
		// if you wanted to inspect the full response data, you could do that here
	})
	if _, err := target.Page.Enable(); err != nil {
	}
	navigateParams := &gcdapi.PageNavigateParams{Url: url}
	_, _, err = target.Page.NavigateWithParams(navigateParams)
	if err != nil {
		panic(err)
	}

	wg.Wait() // wait for page load
}
