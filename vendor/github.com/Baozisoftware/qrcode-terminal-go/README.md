
# qrcode-terminal-go
QRCode terminal for golang.

# Example
```go
package main

import "github.com/Baozisoftware/qrcode-terminal-go"

func main() {
	Test1()
	Test2()
}

func Test1(){
	content := "Hello, 世界"
	obj := qrcodeTerminal.New()
	obj.Get(content).Print()
}

func Test2(){
	content := "https://github.com/Baozisoftware/qrcode-terminal-go"
	obj := qrcodeTerminal.New2(qrcodeTerminal.ConsoleColors.BrightBlue,qrcodeTerminal.ConsoleColors.BrightGreen,qrcodeTerminal.QRCodeRecoveryLevels.Low)
	obj.Get([]byte(content)).Print()
}
```

## Screenshots
### Windows XP
![winxp](https://github.com/Baozisoftware/qrcode-terminal-go/blob/master/screenshots/winxp.png)
### Windows 7
![win7](https://github.com/Baozisoftware/qrcode-terminal-go/blob/master/screenshots/win7.png)
### Windows 10
![win10](https://github.com/Baozisoftware/qrcode-terminal-go/blob/master/screenshots/win10.png)
### Ubuntu
![ubuntu](https://github.com/Baozisoftware/qrcode-terminal-go/blob/master/screenshots/ubuntu.png)
### macOS
![macos](https://github.com/Baozisoftware/qrcode-terminal-go/blob/master/screenshots/macos.png)
