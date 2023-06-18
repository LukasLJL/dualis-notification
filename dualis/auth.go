package dualis

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
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

	if len(resp.Header.Get("REFRESH")) == 0 || resp.StatusCode != 200 {
		log.Fatalln("Could not log in. Check credentials.")
		return nil, false
	} else {
		log.Println("Login successful. Following 1st startup redirect.")
	}

	_, ok = dualis.sessionCookie(resp)
	if !ok {
		log.Fatal("No session cookie configured.")
		return nil, false
	}

	u, _ := url.Parse("https://dualis.dhbw.de")
	dualis.Client.Jar.SetCookies(u, resp.Cookies())
	refreshUrl, _ := dualis.cleanRefreshURL(resp.Header.Get("REFRESH"))
	req, _ = http.NewRequest("GET", baseURL+refreshUrl.String(), nil)
	req.Header.Add("User-Agent", userAgent)
	resp, _ = dualis.Client.Do(req)

	root, _ := html.Parse(resp.Body)
	elem, ok := scrape.Find(root, func(n *html.Node) bool {
		for _, a := range n.Attr {
			return a.Key == "href" && strings.Contains(a.Val, "scripts/mgrqispi.dll")
		}
		return false
	})

	if !ok {
		log.Fatalln("Could not find 2nd startup redirect link.")
		return nil, false
	}

	url, err := url.Parse(elem.Attr[0].Val)

	if err != nil {
		log.Fatalln("Could not parse 2nd startup redirect link.")
		return nil, false
	}

	log.Println("Found 2nd redirect link. Home successfully discovered.")

	return url, true
}
