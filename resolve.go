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
	const (
		recaptchaBase = "https://www.google.com/recaptcha/api2/" // Base URL for reCAPTCHA API
		anchorParams  = "ar=1&k=6Lcr1ncUAAAAAH3cghg6cOTPGARa8adOf-y9zv2x&co=aHR0cHM6Ly9vdW8ucHJlc3M6NDQz&hl=en&v=pCoGBhjs9s8EhFOHJFe8cqis&size=invisible&cb=ahgyd1gkfkhe"
		recaptchaK    = "6Lcr1ncUAAAAAH3cghg6cOTPGARa8adOf-y9zv2x"
		recaptchaV    = "pCoGBhjs9s8EhFOHJFe8cqis"
		recaptchaCo   = "aHR0cHM6Ly9vdW8ucHJlc3M6NDQz"
	)

	client, _ := tlsclient.NewHttpClient(tlsclient.NewNoopLogger())

	anchorURL := recaptchaBase + "anchor?" + anchorParams
	resp, err := client.Do(mustRequest("GET", anchorURL, nil))
	if err != nil {
		return "", fmt.Errorf("anchor request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ := io.ReadAll(resp.Body)
	tokenMatch := regexp.MustCompile(`"recaptcha-token" value="(.*?)"`).FindSubmatch(body)
	if len(tokenMatch) < 2 {
		return "", errors.New("token not found in anchor response")
	}

	postData := url.Values{
		"v":      {recaptchaV},
		"c":      {string(tokenMatch[1])},
		"k":      {recaptchaK},
		"co":     {recaptchaCo},
		"reason": {"q"},
	}

	reloadURL := recaptchaBase + "reload?k=" + recaptchaK
	resp, err = client.Do(mustRequest("POST", reloadURL, strings.NewReader(postData.Encode())))
	if err != nil {
		return "", fmt.Errorf("reload request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ = io.ReadAll(resp.Body)
	if answer := regexp.MustCompile(`"rresp","(.*?)"`).FindSubmatch(body); len(answer) > 1 {
		return string(answer[1]), nil
	}
	return "", errors.New("reCAPTCHA solution not found")
}

// mustRequest helper with standardized headers
func mustRequest(method, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	return req
}
