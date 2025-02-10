package ouo_bypass_go

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// Resolve function to bypass ouo.io and ouo.press links
func Resolve(ouoURL string) (string, error) {
	bypassed, err := OuoBypass(ouoURL)
	if err != nil {
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
	jar := tlsclient.NewCookieJar()
	options := []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(30),
		tlsclient.WithClientProfile(profiles.Chrome_110),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithRandomTLSExtensionOrder(),
		tlsclient.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}

	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	if err != nil {
		return "", err
	}

	getReq, err := http.NewRequest(http.MethodGet, tempURL, nil)
	if err != nil {
		return "", err
	}
	chrome110UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	const accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
	const acceptLang = "en-GB,en-US;q=0.9,en;q=0.8"
	getReq.Header = http.Header{
		"accept":                    {accept},
		"accept-language":           {acceptLang},
		"authority":                 {"ouo.io"},
		"cache-control":             {"max-age=0"},
		"referer":                   {"https://www.google.com/ig/adde?moduleurl="},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {chrome110UserAgent},
		http.HeaderOrderKey: {
			"host",
			"connection",
			"cache-control",
			"authority",
			"accept",
			"accept-language",
			"referer",
			"upgrade-insecure-requests",
		},
	}

	resp, err := client.Do(getReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 403 {
		return "", errors.New("ouo.io is blocking the request")
	}
	readBytes, _ := io.ReadAll(resp.Body)
	data := url.Values{}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(readBytes)))
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		if strings.HasSuffix(name, "token") {
			value, _ := s.Attr("value")
			data.Add(name, value)
		}
	})
	nextURL := fmt.Sprintf("%s://%s/go/%s", u.Scheme, u.Host, id)

	recaptchaV3, err := RecaptchaV3()
	if err != nil {
		return "", err
	}
	data.Set("x-token", recaptchaV3)
	for i := 0; i < 2; i++ {
		time.Sleep(1 * time.Second)
		postReq, err := http.NewRequest(http.MethodPost, nextURL, strings.NewReader(data.Encode()))
		postReq.Header = http.Header{
			"accept":                    {accept},
			"authority":                 {"ouo.io"},
			"cache-control":             {"max-age=0"},
			"content-type":              {"application/x-www-form-urlencoded"},
			"accept-language":           {acceptLang},
			"upgrade-insecure-requests": {"1"},
			"user-agent":                {chrome110UserAgent},
			http.HeaderOrderKey: {
				"host",
				"connection",
				"cache-control",
				"authority",
				"accept",
				"accept-language",
				"referer",
				"upgrade-insecure-requests",
			},
		}
		resp, err := client.Do(postReq)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode == 302 {
			location = resp.Header.Get("Location")
			break
		} else if resp.StatusCode == 403 {
			return "", errors.New("ouo.io is blocking the request")
		}
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
	tempUrl := urlBase + "anchor?" + params
	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger())
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, tempUrl, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("recaptcha status code is not 200")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tokenMatches := regexp.MustCompile(`"recaptcha-token" value="(.*?)"`).FindStringSubmatch(string(body))
	if len(tokenMatches) < 2 {
		return "", errors.New("no token found in response")
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
	postReq, err := http.NewRequest(http.MethodPost, urlBase+"reload?k="+paramsMap["k"], strings.NewReader(postData.Encode()))
	if err != nil {
		return "", err
	}
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(postReq)
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
		return "", errors.New("no answer found in reCaptcha response")
	}

	return answerMatches[1], nil

}
