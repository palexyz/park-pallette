package main

//use this for whatever
import (
	"bytes"
	"fmt"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"image"
	"strconv"

	"gocv.io/x/gocv"
)

type spotVar struct {
	id     int
	status bool
	x      int
	y      int
}

var confirmButton = false
var spots []spotVar
var inputT int
var curFrame image.Image
var spotCount int
var setupFin = false
var findBlocks = false
var contours gocv.PointsVector

func main() {

	fmt.Println("yay")

	webcam, _ := gocv.VideoCaptureDevice(0)
	img := gocv.NewMat()

	hsv := gocv.NewMat()

	mask := gocv.NewMat()

	c1 := make(chan string)
	setup := make(chan bool)

	go func() {
		temp := true
		for { //This is the webcam

			webcam.Read(&img)

			if setupFin {
				//H: 147 S: 189 V: 196
				//bgr 196 51 184

				iRows, iColumns := img.Rows(), img.Cols()
				lower := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(137, 100, 100, 0.0), iRows, iColumns, gocv.MatTypeCV8UC3)
				upper := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(157, 255, 255, 0.0), iRows, iColumns, gocv.MatTypeCV8UC3)

				gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)

				gocv.InRange(hsv, lower, upper, &mask)

				kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
				gocv.Dilate(mask, &mask, kernel)

				if temp {
					contours = gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
					for g := 0; g < inputT; g++ {

						spots[g].x = gocv.MinAreaRect(contours.At(g)).Center.X
						spots[g].y = gocv.MinAreaRect(contours.At(g)).Center.Y
						fmt.Println(spots[g])

					}
					temp = false
				}

				for i := 0; i < spotCount; i++ {
					tempMat := mask.Region(image.Rect(spots[i].x-1, spots[i].y-1, spots[i].x, spots[i].y))
					tempI := tempMat.ToBytes()

					ahsbsh := []byte{0}

					if bytes.Equal(tempI, ahsbsh) {
						spots[i].status = false

					} else {
						spots[i].status = true
					}

					if spots[i].status {
						gocv.Circle(&img, image.Pt(spots[i].x, spots[i].y), 10, color.RGBA{0, 255, 0, 0}, 3)

					} else {
						gocv.Circle(&img, image.Pt(spots[i].x, spots[i].y), 10, color.RGBA{255, 0, 0, 0}, 3)
						c1 <- " "

					}

				}

				gocv.DrawContours(&img, contours, -1, color.RGBA{0, 255, 0, 255}, 1)

			}
			curFrame, _ = img.ToImage()

		}
	}()

	a := app.New()
	w := a.NewWindow("Park Pallette")
	w.Resize(fyne.NewSize(750, 350))

	input := widget.NewEntry()
	inputCon := container.NewVBox(input)
	inputCon.MinSize()

	input.SetPlaceHolder("Number of spots")

	button := widget.NewButton("confirm", func() {
		inputT, _ = strconv.Atoi(input.Text)
		if inputT > 0 {
			setup <- true
		} else {
			input.SetPlaceHolder("Invalid")
		}
	})

	grid := container.NewGridWithRows(2, inputCon, button)

	content := container.NewGridWithColumns(1, grid)

	w.SetContent(content)

	go func() {
		spotText := "" // the thing to check if setup finishes

		for i := 0; i <= 0; {
			select {
			case setupVal := <-setup: //checks for status of setup
				fmt.Println("Setup:", inputT, setupVal)
				spotCount = inputT
				for l := 1; l <= inputT; l++ { //makes spot variables
					var s spotVar

					s.id = l
					spots = append(spots, s)
				}
				for v := 1; v <= inputT; v++ {
					spotText += ("\nID: " + strconv.Itoa(spots[v-1].id) + "\n -Active: " + strconv.FormatBool(spots[v-1].status) + "\n -Coordinates: " + strconv.Itoa(spots[v-1].x) + "," + strconv.Itoa(spots[v-1].y) + "\n")
				}
				fmt.Println(spots)

				text := container.NewCenter(widget.NewLabel(spotText))
				image := canvas.NewImageFromImage(curFrame)
				image.SetMinSize(fyne.NewSize(200, 200))
				image.FillMode = canvas.ImageFillContain

				content = container.NewGridWithColumns(2, image, text)

				w.SetContent(content)

				setupFin = true
				i = 200
			}
		}
	}()

	go func() { //this is the screen after setup
		//pretty much the same as the one above
		for {
			if setupFin {
				spotText := ""

				select {
				case wow := <-c1:

					for v := 1; v <= inputT; v++ {
						spotText += ("\nID:" + wow + strconv.Itoa(spots[v-1].id) + "\n -Active: " + strconv.FormatBool(spots[v-1].status) + "\n -Coordinates: " + strconv.Itoa(spots[v-1].x) + "," + strconv.Itoa(spots[v-1].y) + "\n")
					}

					text := container.NewCenter(widget.NewLabel(spotText))
					image := canvas.NewImageFromImage(curFrame)
					image.SetMinSize(fyne.NewSize(200, 200))
					image.FillMode = canvas.ImageFillContain

					content = container.NewGridWithColumns(2, image, text)

					w.SetContent(content)
					//go callme(spots)
				}
			}
		}
	}()

	w.ShowAndRun()
	os.Exit(1)

}
