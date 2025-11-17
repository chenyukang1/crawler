package utils

import "net/url"

func UrlEncode(s string) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	u.RawQuery = u.Query().Encode()
	return u, nil
}
