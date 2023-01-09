package dualis

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (dualis *Dualis) login(username, password string) (homeUrl *url.URL, ok bool) {
	postData := url.Values{"usrname": {username},
		"pass":      {password},
		"APPNAME":   {"CampusNet"},
		"PRGNAME":   {"LOGINCHECK"},
		"ARGUMENTS": {"clino,usrname,pass,menuno,menu_type,browser,platform"},
		"clino":     {"000000000000001"},
		"menuno":    {"000324"},
		"menu_type": {"classic"},
		"browser":   {""},
		"platform":  {""},
	}

	req, _ := http.NewRequest("POST", baseURL+loginScriptPath, strings.NewReader(postData.Encode()))
	req.Header.Add("User-Agent", userAgent)

	resp, _ := dualis.Client.Do(req)

	if len(resp.Header.Get("REFRESH")) == 0 {
		log.Fatalln("Could not log in. Check credentials.")
		var u *url.URL
		return u, false
	} else {
		log.Println("Login successful. Following 1st startup redirect.")
	}

	_, ok = dualis.sessionCookie(resp)
	if !ok {
		log.Fatal("No session cookie configured.")
		var u *url.URL
		return u, false
	}

	u, _ := url.Parse("https://dualis.dhbw.de")
	dualis.Client.Jar.SetCookies(u, resp.Cookies())
	refreshUrl, _ := dualis.cleanRefreshURL(resp.Header.Get("REFRESH"))
	req, _ = http.NewRequest("GET", baseURL+refreshUrl.String(), nil)
	req.Header.Add("User-Agent", userAgent)
	resp, _ = dualis.Client.Do(req)

	root, _ := html.Parse(resp.Body)
	elem, ok := scrape.Find(root, func(n *html.Node) bool {
		return n.DataAtom == atom.Meta && n.Attr[0].Key == "http-equiv" && n.Attr[0].Val == "refresh"
	})

	if !ok {
		log.Fatalln("Could not find 2nd startup redirect link.")
		var u *url.URL
		return u, false
	}

	redirectUrl, _ := dualis.cleanRefreshURL(elem.Attr[1].Val)
	log.Println("Found 2nd redirect link. Home successfully discovered.")

	return redirectUrl, true
}
