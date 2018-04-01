// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element.

// used as processor
// # Inputs :
// - html content (field name)
// - ids of elements to image
// - StripCSSANDJSLink

// # output :
// same event with html content replaced with html with embeded images or local images

// SingleHTMLStatic{
//   fieldName => "content"
//   FixIds =>
//   EmbedasB64 =>
//   RemoveScript => css, js
// }

package main

import (
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/clbanning/mxj"
	xp "github.com/vjeantet/bitfan/commons/xprocessor"
)

var r *xp.Runner

/*
{
    "@timestamp": "2018-02-26T21:02:33.687343+01:00",
    "message": "{\"message\":\"Hello world\"}",
    "output": "\u003c!DOCTYPE HTML PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional //EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\"\u003e\r\n\u003chtml xmlns=\"http://www.w3.org/1999/xhtml\"\u003e\r\n\u003chead\u003e\r\n    \u003ctitle\u003eEmail\u003c/title\u003e\r\n    \r\n         \u003c!--Import Google Icon Font--\u003e\r\n      \u003c!--Import materialize.css--\u003e\r\n      \r\n\r\n    \r\n    \r\n\u003c/head\u003e\r\n\u003cbody\u003e\r\n  \r\n      \r\n      \u003cnav id=\"navbar\"\u003e\r\n    \u003cdiv class=\"nav-wrapper\"\u003e\r\n      \u003ca href=\"#\" class=\"brand-logo\"\u003eLogo\u003c/a\u003e\r\n      \u003cul id=\"nav-mobile\" class=\"right hide-on-med-and-down\"\u003e\r\n        \u003cli\u003e\u003ca href=\"sass.html\"\u003eSass\u003c/a\u003e\u003c/li\u003e\r\n        \u003cli\u003e\u003ca href=\"badges.html\"\u003eComponents\u003c/a\u003e\u003c/li\u003e\r\n        \u003cli\u003e\u003ca href=\"collapsible.html\"\u003eJavaScript\u003c/a\u003e\u003c/li\u003e\r\n      \u003c/ul\u003e\r\n    \u003c/div\u003e\r\n  \u003c/nav\u003e\r\n    \r\n   \u003ch1\u003eTest mail\u003c/h1\u003e\r\n   \u003cp\u003e{\"message\":\"Hello world\"}\u003c/p\u003e\r\n   Embeded image !\r\n   \u003cimg src=\"http://127.0.0.1:5123/public/oss.png\" alt=\"My image\" /\u003e\r\n   \r\n   Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod\r\ntempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,\r\nquis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo\r\nconsequat. Duis aute irure dolor in reprehenderit in voluptate velit esse\r\ncillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non\r\nproident, sunt in culpa qui officia deserunt mollit anim id est laborum.\r\n   \r\n   \u003cdiv id=test style=\"width: 120px; height: 150px; background-color: red; color: white; text-align: center\"\u003e\r\n\t\u003cdiv style=\"font-size: 50px;padding-top: 20px\"\u003e110\u003c/div\u003e\r\n\t\u003cdiv style=\"font-size: 10px;color: white\"\u003eLe nombre vient de JIRA\u003c/div\u003e\r\n   \u003c/div\u003e\r\n   \r\nLorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod\r\ntempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,\r\nquis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo\r\nconsequat. Duis aute irure dolor in reprehenderit in voluptate velit esse\r\ncillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non\r\nproident, sunt in culpa qui officia deserunt mollit anim id est laborum.\r\n\r\n\u003c/body\u003e\r\n\u003c/html\u003e"
}
*/

func main() {
	r = xp.New(Configure, Start, Receive, Stop)
	r.Description = "modify html with embeded images in place of elements"
	r.ShortDescription = ""

	r.OptionString("source", true, "source fieldname", "output")
	r.OptionStringSlice("selectors", true, "selectors to html elements to replace by an image", nil)
	r.OptionBool("show_browser", false, "Show browser whilte processing", false)
	r.OptionStringSlice("remove_selectors", false, "Remove elements found with selectors", nil)

	r.Run(1)
}

func Configure() error {
	return nil
}

var cs *ChromeShot

func Start() error {
	var err error
	cs, err = NewChromeShot(r.Opt.Bool("show_browser"), r.Debugf)
	if err != nil {
		r.Logf("error ", err.Error())
		os.Exit(2)
	}
	return nil
}
func Receive(data interface{}) error {
	var err error
	message := mxj.Map(data.(map[string]interface{}))
	htmlContent := message.ValueOrEmptyForPathString(r.Opt.String("source"))
	// r.Debugf("%s", htmlContent)

	htmlContent, err = cs.EmbedImageForDomElements(htmlContent, r.Opt.StringSlice("selectors"))
	if err != nil {
		r.Logf(err.Error())
		return err
	}

	if len(r.Opt.StringSlice("remove_selectors")) > 0 {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		for _, selector := range r.Opt.StringSlice("remove_selectors") {
			sel := doc.Find(selector)
			for k := range sel.Nodes {
				sel.Eq(k).Remove()
			}
		}
		htmlContent, _ = doc.Html()
	}

	message.SetValueForPath(htmlContent, r.Opt.String("source"))

	return r.Send(message)
}
func Stop() error {
	cs.Stop()
	return nil
}
