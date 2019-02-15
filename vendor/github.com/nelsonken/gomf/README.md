# golang 可多文件上传的request builder 库

## 测试方法

1. start php upload server: php -S 127.0.0.1:8080 ./
2. run go test -v 

## 使用方法

```go
	fb := gomf.New()
	fb.WriteField("name", "accountName")
	fb.WriteField("password", "pwd")
	fb.WriteFile("picture", "icon.png", "image/jpeg", []byte(strings.Repeat("0", 100)))

	log.Println(fb.GetBuffer().String())

	req, err := fb.GetHTTPRequest(context.Background(), "http://127.0.0.1:8080/up.php")
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)

	log.Println(res.StatusCode)
	log.Println(res.Status)

	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(b))
```
