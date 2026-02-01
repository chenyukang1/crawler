package spider

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"

	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/pkg/log"
)

type Context struct {
	Spider         *Spider
	URL            string
	Request        *http.Request
	Response       *http.Response
	StructuredData []collect.DataCell

	html []byte            // html二进制数据
	dom  *goquery.Document // 解析dom节点
}

type RuleFunc func(ctx *Context)

func (c *Context) Rule(ruleName string) error {
	if c.Spider.Rules[ruleName] == nil {
		return errors.New("rule " + ruleName + " not found")
	}
	c.Spider.Rules[ruleName].Run(c)
	return nil
}

func (c *Context) GetHtml() ([]byte, error) {
	if c.html == nil {
		if err := c.parseHTML(); err != nil {
			log.Errorf("parse html from %s fail %v", c.URL, err)
			return nil, err
		}
	}
	return c.html, nil
}

func (c *Context) GetDom() (dom *goquery.Document, err error) {
	if c.dom == nil {
		var html []byte
		html, err = c.GetHtml()
		if err != nil {
			return nil, err
		}
		dom, err = goquery.NewDocumentFromReader(bytes.NewReader(html))
		if err != nil {
			log.Errorf("parse dom from %s fail %v", html, err)
			return nil, err
		}
		return
	}
	return c.dom, nil
}

func (c *Context) parseHTML() error {
	var contentType, pageEncode string
	// 优先从响应头读取编码类型
	contentType = c.Response.Header.Get("Content-Type")
	if _, params, err := mime.ParseMediaType(contentType); err == nil {
		if cs, ok := params["charset"]; ok {
			pageEncode = strings.ToLower(strings.TrimSpace(cs))
		}
	}
	// 响应头未指定编码类型时，从请求头读取
	if len(pageEncode) == 0 {
		contentType = c.Request.Header.Get("Content-Type")
		if _, params, err := mime.ParseMediaType(contentType); err == nil {
			if cs, ok := params["charset"]; ok {
				pageEncode = strings.ToLower(strings.TrimSpace(cs))
			}
		}
	}
	defer func() {
		if err := c.Response.Body.Close(); err != nil {
			log.Warnf("【%s】响应body关闭失败!!", c.URL)
		}
	}()

	var err error
	switch pageEncode {
	case "utf8", "utf-8", "unicode-1-1-utf-8":
		// 不做转码处理
		c.html, err = ioutil.ReadAll(c.Response.Body)
	default:
		// 指定了编码类型，但不是utf8时，自动转码为utf8
		// get converter to utf-8
		// Charset auto determine. Use golang.org/x/net/html/charset. Get response body and change it to utf-8
		var destReader io.Reader
		if len(pageEncode) == 0 {
			destReader, err = charset.NewReader(c.Response.Body, "")
		} else {
			destReader, err = charset.NewReaderLabel(pageEncode, c.Response.Body)
		}

		if err == nil {
			c.html, err = ioutil.ReadAll(destReader)
		}
	}
	return err
}
