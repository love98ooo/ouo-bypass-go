package ouo_bypass_go

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Resolve function to bypass ouo.io and ouo.press links
func Resolve(ouoURL string) (string, error) {
	bypassed, err := OuoBypass(ouoURL)
	if err != nil {
		log.Default().Println("Resolve ouo short-link Error: ", err)
		return "", err
	}
	return bypassed, nil
}

// OuoBypass function to bypass ouo.io and ouo.press links
func OuoBypass(ouoURL string) (string, error) {
	tempURL := strings.Replace(ouoURL, "ouo.press", "ouo.io", 1)
	var location string
	u, err := url.Parse(tempURL)
	if err != nil {
		return "", err
	}
	id := tempURL[strings.LastIndex(tempURL, "/")+1:]
	ja3Transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := colly.NewCollector()
	client.WithTransport(ja3Transport)
	client.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		location = req.URL.String()
		return http.ErrUseLastResponse
	})
	extensions.RandomUserAgent(client)
	data := make(map[string]string)
	client.OnResponse(func(r *colly.Response) {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(r.Body)))
		doc.Find("input").Each(func(i int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if strings.HasSuffix(name, "token") {
				data[name], _ = s.Attr("value")
			}
		})
	})
	_ = client.Visit(tempURL)
	nextURL := fmt.Sprintf("%s://%s/go/%s", u.Scheme, u.Host, id)

	for i := 0; i < 2; i++ {
		log.Default().Println("ouo short-link next URL: ", nextURL)
		recaptchaV3, err := RecaptchaV3()
		if err != nil {
			return "", err
		}

		data["x-token"] = recaptchaV3
		err = client.Post(nextURL, data)
		if errors.Is(err, http.ErrUseLastResponse) {
			return location, nil
		}
		//client.OnResponse(func(r *colly.Response) {
		//	fmt.Println(r.StatusCode)
		//})
		nextURL = fmt.Sprintf("%s://%s/xreallcygo/%s", u.Scheme, u.Host, id)
	}
	return location, nil
}

// RecaptchaV3 function to bypass reCAPTCHA v3
func RecaptchaV3() (string, error) {
	AnchorUrl := "https://www.google.com/recaptcha/api2/anchor?ar=1&k=6Lcr1ncUAAAAAH3cghg6cOTPGARa8adOf-y9zv2x&co=aHR0cHM6Ly9vdW8ucHJlc3M6NDQz&hl=en&v=pCoGBhjs9s8EhFOHJFe8cqis&size=invisible&cb=ahgyd1gkfkhe"
	urlBase := "https://www.google.com/recaptcha/"

	matches := regexp.MustCompile(`([api2|enterprise]+)\/anchor\?(.*)`).FindStringSubmatch(AnchorUrl)
	if len(matches) < 3 {
		return "", fmt.Errorf("no matches found in ANCHOR_URL")
	}

	urlBase += matches[1] + "/"
	params := matches[2]

	resp, err := http.Get(urlBase + "anchor?" + params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tokenMatches := regexp.MustCompile(`"recaptcha-token" value="(.*?)"`).FindStringSubmatch(string(body))
	if len(tokenMatches) < 2 {
		return "", fmt.Errorf("no token found in response")
	}
	token := tokenMatches[1]
	paramsMap := make(map[string]string)
	for _, pair := range strings.Split(params, "&") {
		parts := strings.Split(pair, "=")
		if len(parts) == 2 {
			paramsMap[parts[0]] = parts[1]
		}
	}
	postData := url.Values{}
	postData.Set("v", paramsMap["v"])
	postData.Set("c", token)
	postData.Set("k", paramsMap["k"])
	postData.Set("co", paramsMap["co"])
	postData.Set("reason", "q")

	resp, err = http.Post(urlBase+"reload?k="+paramsMap["k"], "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	answerMatches := regexp.MustCompile(`"rresp","(.*?)"`).FindStringSubmatch(string(body))
	if len(answerMatches) < 2 {
		return "", fmt.Errorf("no answer found in response")
	}

	return answerMatches[1], nil

}
