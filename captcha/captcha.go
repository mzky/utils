package captcha

import (
	_ "embed" // embed font
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const charPreset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

//go:embed fonts/Comismsh.ttf
var ttf []byte
var ttfFont *truetype.Font

// Options manage captcha generation details.
type Options struct {
	// BackgroundColor is captcha image's background color.
	// It defaults to color.Transparent.
	BackgroundColor color.Color
	// CharPreset decides what text will be on captcha image.
	// It defaults to digit 0-9 and all English alphabet.
	// ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789
	CharPreset string
	// TextLength is the length of captcha text.
	// It defaults to 4.
	TextLength int
	// CurveNumber is the number of curves to draw on captcha image.
	// It defaults to 2.
	CurveNumber int
	// FontDPI controls DPI (dots per inch) of font.
	// The default is 72.0.
	FontDPI float64
	// FontScale controls the scale of font.
	// The default is 1.0.
	FontScale float64
	// Noise controls the number of noise drawn.
	// A noise dot is drawn for every 28 pixel by default.
	// The default is 1.0.
	Noise float64
	// Palette is the set of colors to chose from
	Palette color.Palette

	width  int
	height int
}

func newDefaultOption(width, height int) *Options {
	return &Options{
		BackgroundColor: color.Transparent,
		CharPreset:      charPreset,
		TextLength:      4,
		CurveNumber:     2,
		FontDPI:         72.0,
		FontScale:       1.0,
		Noise:           1.0,
		Palette:         []color.Color{},
		width:           width,
		height:          height,
	}
}

// SetOption is a function that can be used to modify default options.
type SetOption func(*Options)

// Data is the result of captcha generation.
// It has a `Text` field and a private `img` field that will
// be used in `WriteImage` receiver.
type Data struct {
	// Text is captcha solution.
	Text string
	Img  *image.NRGBA
}

// WriteImage encodes image data and writes to an io.Writer.
// It returns possible error from PNG encoding.
func (data *Data) WriteImage(w io.Writer) error {
	return png.Encode(w, data.Img)
}

// WriteJPG encodes image data in JPEG format and writes to an io.Writer.
// It returns possible error from JPEG encoding.
func (data *Data) WriteJPG(w io.Writer, o *jpeg.Options) error {
	return jpeg.Encode(w, data.Img, o)
}

// WriteGIF encodes image data in GIF format and writes to an io.Writer.
// It returns possible error from GIF encoding.
func (data *Data) WriteGIF(w io.Writer, o *gif.Options) error {
	return gif.Encode(w, data.Img, o)
}

// WritePNGFile 将图像保存为PNG文件
func (data *Data) WritePNGFile(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, data.Img)
}

// WriteJPGFile 将图像保存为JPG文件
func (data *Data) WriteJPGFile(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, data.Img, &jpeg.Options{Quality: 75})
}

// WriteGIFFile 将图像保存为GIF文件
func (data *Data) WriteGIFFile(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	return gif.Encode(f, data.Img, &gif.Options{})
}

func init() {
	ttfFont, _ = freetype.ParseFont(ttf)
	rand.Seed(time.Now().UnixNano())
}

// LoadFont let you load an external font.
func LoadFont(fontData []byte) error {
	var err error
	ttfFont, err = freetype.ParseFont(fontData)
	return err
}

// LoadFontFromReader load an external font from an io.Reader interface.
func LoadFontFromReader(reader io.Reader) error {
	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return LoadFont(b)
}

// New creates a new captcha.
// It returns captcha data and any freetype drawing error encountered.
func New(width int, height int, option ...SetOption) (*Data, error) {
	options := newDefaultOption(width, height)
	for _, setOption := range option {
		setOption(options)
	}

	text := randomText(options)
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	if err := drawWithOption(text, img, options); err != nil {
		return nil, err
	}

	return &Data{Text: text, Img: img}, nil
}

// NewMathExpr creates a new captcha.
// It will generate a image with a math expression like `1 + 2`.
func NewMathExpr(width int, height int, option ...SetOption) (*Data, error) {
	options := newDefaultOption(width, height)
	for _, setOption := range option {
		setOption(options)
	}

	text, equation := randomEquation()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	if err := drawWithOption(equation, img, options); err != nil {
		return nil, err
	}

	return &Data{Text: text, Img: img}, nil
}

// NewCustomGenerator creates a new captcha based on a custom text generator.
func NewCustomGenerator(
	width int, height int, generator func() (anwser string, question string), option ...SetOption,
) (*Data, error) {
	options := newDefaultOption(width, height)
	for _, setOption := range option {
		setOption(options)
	}

	answer, question := generator()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	if err := drawWithOption(question, img, options); err != nil {
		return nil, err
	}

	return &Data{Text: answer, Img: img}, nil
}

func drawWithOption(text string, img *image.NRGBA, options *Options) error {
	draw.Draw(img, img.Bounds(), &image.Uniform{C: options.BackgroundColor}, image.Point{}, draw.Src)
	drawNoise(img, options)
	drawCurves(img, options)
	return drawText(text, img, options)
}

func randomText(opts *Options) (text string) {
	n := len([]rune(opts.CharPreset))
	for i := 0; i < opts.TextLength; i++ {
		text += string([]rune(opts.CharPreset)[rand.Intn(n)])
	}

	return text
}

func drawNoise(img *image.NRGBA, opts *Options) {
	noiseCount := (opts.width * opts.height) / int(28.0/opts.Noise)
	for i := 0; i < noiseCount; i++ {
		x := rand.Intn(opts.width)
		y := rand.Intn(opts.height)
		img.Set(x, y, randomColor())
	}
}

func randomColor() color.RGBA {
	red := rand.Intn(256)
	green := rand.Intn(256)
	blue := rand.Intn(256)

	return color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: uint8(255)}
}

func drawCurves(img *image.NRGBA, opts *Options) {
	for i := 0; i < opts.CurveNumber; i++ {
		drawSineCurve(img, opts)
	}
}

// Ideally we want to draw bezier curves
// For now sine curves will do the job
func drawSineCurve(img *image.NRGBA, opts *Options) {
	var xStart, xEnd int
	if opts.width <= 30 {
		xStart, xEnd = 1, opts.width-1
	} else {
		xStart = rand.Intn(opts.width/10) + 1
		xEnd = opts.width - rand.Intn(opts.width/10) - 1
	}
	curveHeight := float64(rand.Intn(opts.height/6) + opts.height/6)
	yStart := rand.Intn(opts.height*2/3) + opts.height/6
	angle := 1.0 + rand.Float64()
	yFlip := 1.0
	if rand.Intn(2) == 0 {
		yFlip = -1.0
	}
	curveColor := randomColorFromOptions(opts)

	for x1 := xStart; x1 <= xEnd; x1++ {
		y := math.Sin(math.Pi*angle*float64(x1)/float64(opts.width)) * curveHeight * yFlip
		img.Set(x1, int(y)+yStart, curveColor)
	}
}

func drawText(text string, img *image.NRGBA, opts *Options) error {
	ctx := freetype.NewContext()
	ctx.SetDPI(opts.FontDPI)
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetHinting(font.HintingFull)
	ctx.SetFont(ttfFont)

	fontSpacing := opts.width/len(text) - 2
	fontOffset := rand.Intn(fontSpacing / 2)

	for idx, char := range text {
		fontScale := 0.8 + rand.Float64()*0.5
		fontSize := float64(opts.height) / fontScale * opts.FontScale
		ctx.SetFontSize(fontSize)
		ctx.SetSrc(image.NewUniform(randomColorFromOptions(opts)))
		x := fontSpacing*idx + fontOffset
		y := opts.height/10 + rand.Intn(opts.height/3) + int(fontSize/2)
		pt := freetype.Pt(x, y)
		if _, err := ctx.DrawString(string(char), pt); err != nil {
			return err
		}
	}

	return nil
}

func randomColorFromOptions(opts *Options) color.Color {
	length := len(opts.Palette)
	if length == 0 {
		return randomInvertColor(opts.BackgroundColor)
	}

	return opts.Palette[rand.Intn(length)]
}

func randomInvertColor(base color.Color) color.Color {
	baseLightness := getLightness(base)
	var value float64
	if baseLightness >= 0.5 {
		value = baseLightness - 0.3 - rand.Float64()*0.2
	} else {
		value = baseLightness + 0.3 + rand.Float64()*0.2
	}
	hue := float64(rand.Intn(361)) / 360
	saturation := 0.6 + rand.Float64()*0.2

	return hsva{h: hue, s: saturation, v: value, a: 255}
}

func getLightness(colour color.Color) float64 {
	r, g, b, a := colour.RGBA()
	// transparent
	if a == 0 {
		return 1.0
	}

	l := (float64(maxColor(r, g, b)) + float64(minColor(r, g, b))) / (2 * 255)

	return l
}

func maxColor(numList ...uint32) (max uint32) {
	for _, num := range numList {
		colorVal := num & 255
		if colorVal > max {
			max = colorVal
		}
	}

	return max
}

func minColor(numList ...uint32) (min uint32) {
	min = 255
	for _, num := range numList {
		colorVal := num & 255
		if colorVal < min {
			min = colorVal
		}
	}

	return min
}

func randomEquation() (text string, equation string) {
	left := 1 + rand.Intn(9)
	right := 1 + rand.Intn(9)
	text = strconv.Itoa(left + right)
	equation = strconv.Itoa(left) + "+" + strconv.Itoa(right)

	return text, equation
}
