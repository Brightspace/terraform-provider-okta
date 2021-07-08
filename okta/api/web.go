package api

import (
	"io"

	"golang.org/x/net/html"
)

func getXsrfToken(resBody io.Reader) string {
	z := html.NewTokenizer(resBody)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return ""
		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "span" {
				for _, a := range t.Attr {
					if a.Key == "id" && a.Val == "_xsrfToken" {
						z.Next()
						return z.Token().Data
					}
				}
			}
		}
	}
}
