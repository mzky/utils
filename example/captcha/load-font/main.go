package main

import (
	"fmt"
	"github.com/mzky/utils/captcha"
	"golang.org/x/image/font/gofont/gobold"
	"html/template"
	"image/color"
	"net/http"
)

func main() {
	captchaImg()
	//err := captcha.LoadFont(goregular.TTF)
	//if err != nil {
	//	panic(err)
	//}
	//
	//http.HandleFunc("/", indexHandle)
	//http.HandleFunc("/captcha", captchaHandle)
	//fmt.Println("Server start at port 8080")
	//err = http.ListenAndServe(":8080", nil)
	//if err != nil {
	//	panic(err)
	//}
}

func indexHandle(w http.ResponseWriter, _ *http.Request) {
	doc, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	doc.Execute(w, nil)
}

func captchaHandle(w http.ResponseWriter, _ *http.Request) {
	img, err := captcha.New(120, 35, func(options *captcha.Options) {
		options.FontScale = 0.8
	})
	if err != nil {
		fmt.Fprint(w, nil)
		fmt.Println(err.Error())
		return
	}
	img.WriteImage(w)
}

func captchaImg() {
	err := captcha.LoadFont(gobold.TTF)
	if err != nil {
		panic(err)
	}
	img, _ := captcha.New(100, 35, func(options *captcha.Options) {
		options.CharPreset = "PQRWgvwMm"
		options.FontDPI = 50
		options.Noise = 0.6
		options.BackgroundColor = color.Opaque
	})

	//img.WriteGIFFile("output.gif")
	//img.WriteJPGFile("output.jpg")
	img.WritePNGFile("output.png")

}
