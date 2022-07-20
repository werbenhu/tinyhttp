package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/gin-gonic/gin"
)

var p = flag.String("p", "", "port")

func OpenBrowser(uri string) error {
	cmd := exec.Command(`cmd`, `/c`, `start`, uri)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
	return nil
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func start() {
	r := gin.Default()
	r.Use(cors())

	r.Static("/", ".")

	err := r.Run(":" + *p)
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
}

func onReady() {
	ico, err := ioutil.ReadFile("./tinyhttp.ico")
	if err != nil {
		panic(err)
	}

	systray.SetIcon(ico)
	systray.SetTitle("tinyhttp")
	systray.SetTooltip("Right click to open the menu")
	mShow := systray.AddMenuItem("Show Log", "Show Log")
	mHide := systray.AddMenuItem("Hide Log", "Hide Log")
	admin := systray.AddMenuItem("Open Homepage	", "Open Homepage")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Exit", "Exit")

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	// https://docs.microsoft.com/en-us/windows/console/getconsolewindow
	getConsoleWindows := kernel32.NewProc("GetConsoleWindow")
	// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-showwindowasync
	showWindowAsync := user32.NewProc("ShowWindowAsync")
	consoleHandle, r2, err := getConsoleWindows.Call()
	if consoleHandle == 0 {
		fmt.Println("Error call GetConsoleWindow: ", consoleHandle, r2, err)
	}
	showWindowAsync.Call(consoleHandle, 0)
	// showWindowAsync.Call(consoleHandle, 5)

	adminUrl := "http://localhost:" + *p

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				mShow.Disable()
				mHide.Enable()
				r1, r2, err := showWindowAsync.Call(consoleHandle, 5)
				if r1 != 1 {
					fmt.Println("Error call ShowWindow @SW_SHOW: ", r1, r2, err)
				}
			case <-mHide.ClickedCh:
				mHide.Disable()
				mShow.Enable()
				r1, r2, err := showWindowAsync.Call(consoleHandle, 0)
				if r1 != 1 {
					fmt.Println("Error call ShowWindow @SW_HIDE: ", r1, r2, err)
				}
			case <-admin.ClickedCh:
				OpenBrowser(adminUrl)
			case <-quit.ClickedCh:
				fmt.Println("Quit Server!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				systray.Quit()
			}
		}
	}()

	go start()
	OpenBrowser(adminUrl)

}

func onExit() {
	fmt.Printf("Exit!!!")
}

func main() {

	flag.Parse()
	if *p == "" {
		*p = "8080"
	}
	systray.Run(onReady, onExit)
}
