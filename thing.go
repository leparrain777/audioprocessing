package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"slices"

	"github.com/tosone/minimp3"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type Ordered interface {
	int | int16 | float32 | float64
}

func Max[T Ordered](x, y T) T {
	return T(math.Max(float64(x), float64(y)))
}

func Min[T Ordered](x, y T) T {
	return T(math.Min(float64(x), float64(y)))
}

func Ceil[T Ordered](x T) T {
	return T(math.Ceil(float64(x)))
}

// notes 2^(1/12)=1.059 plus a bit 1.05946309436
// for one cycle distinction we solve 2^(1/12)=(x+1)/x for x with a result of 16.95 or 17 cycles
// I am going to use this as our base time

// one cycle a second is 2PIx

func windowcalc(checkagainst []int16, note float32, samplerate int, indextouse int, combined float64, noteindex int, done chan []int) {

	var lengthofwindow int = int(math.Ceil(combined))
	if lengthofwindow%2 == 0 {
		lengthofwindow++
	}

	windowindex := make([]int16, lengthofwindow)
	var lengthofside int = int(math.Floor(float64(lengthofwindow) / 2))
	for i, _ := range windowindex {
		windowindex[i] = -int16(lengthofside) + int16(i)
	}
	sample := make([]float64, lengthofwindow)
	for i, val := range windowindex {
		sample[i] = math.Cos(combined * float64(val))
	}

	var sum float64 = 0
	if lengthofwindow*2 >= len(checkagainst) {
		sum = 0
	} else if indextouse <= lengthofside {
		var bound int = lengthofside - indextouse
		for i, val := range sample[bound:] {
			sum += float64(checkagainst[i]) * val
		}
	} else if indextouse-lengthofside+lengthofwindow >= len(checkagainst) {
		var bound int = len(checkagainst) - indextouse
		for i, val := range sample[:lengthofwindow-bound] {
			sum += float64(checkagainst[indextouse-(lengthofwindow-bound)+i]) * val
		}
	} else {
		for i, val := range sample {
			sum += float64(checkagainst[indextouse-lengthofside+i-1]) * val
		}
	}
	msg := make([]int, 0)
	msg = append(msg, noteindex, indextouse, int(sum/float64(lengthofwindow)))
	done <- msg
	//return int16(sum / float64(lengthofwindow))
}

func dotransform(d []int16, n []float32, sr int) [][]int16 {

	a := make([][]int16, len(n))
	for ran := range a {
		a[ran] = make([]int16, len(d))
	}

	done := make(chan []int)

	for i, note := range n[:] {
		fmt.Println(i)
		var combined float64 = 2 * math.Pi * 8 / float64(note) * float64(sr)
		fmt.Println(combined)
		var interval int = 1000
		var start int = 0
		var end int = interval
		for l := 0; l < Ceil(len(d)/interval); l++ {
			for j, _ := range d[start:Min(end, len(d))] {
				//a[i][j] = windowcalc(d, note, sr, j, combined)
				go windowcalc(d, note, sr, start+j, combined, i, done)
			}
			//holder := make([]int,3)
			for range d[start:Min(end, len(d))] {
				holder := <-done
				//fmt.Println(holder)
				a[i][holder[1]] = int16(holder[2])
			}
			start += interval
			end += interval
		}
	}
	return a
}

func main() {
	var err error

	var notes []float32
	notes = make([]float32, 0)
	notes = append(notes, 8.18, 8.66, 9.18, 9.72, 10.3, 10.91, 11.56, 12.25, 12.98, 13.75)
	notes = append(notes, 14.57, 15.43, 16.35, 17.32, 18.35, 19.45, 20.6, 21.83, 23.12, 24.5, 25.96, 27.50)
	notes = append(notes, 29.14, 30.87, 32.7, 34.65, 36.71, 38.89, 41.2, 43.65, 46.25, 49, 51.91, 55)
	notes = append(notes, 58.27, 61.74, 65.41, 69.3, 73.42, 77.78, 82.41, 87.31, 92.5, 98, 103.83, 110)
	notes = append(notes, 116.54, 123.47, 130.81, 138.59, 146.83, 155.56, 164.81, 174.61, 185.0, 196.0, 207.65, 220)
	notes = append(notes, 233.08, 246.94, 261.63, 277.18, 293.66, 311.13, 329.63, 349.23, 369.99, 392, 415.3, 440)
	notes = append(notes, 466.16, 493.88, 523.25, 554.37, 587.33, 622.25, 659.26, 698.46, 739.99, 783.99, 830.61, 880)
	notes = append(notes, 932.33, 987.77, 1046.5, 1108.73, 1174.66, 1244.51, 1318.51, 1396.91, 1479.98, 1567.98, 1661.22, 1760)
	notes = append(notes, 1864.66, 1975.53, 2093, 2217.46, 2349.32, 2489.02, 2637.02, 2793.83, 2959.96, 3135.96, 3322.44, 3520)
	notes = append(notes, 3729.31, 3951.07, 4186.01, 4434.92, 4698.64, 4978.03, 5274.04, 5587.65, 5919.91, 6271.93, 6644.88, 7040)
	notes = append(notes, 7458.62, 7902.13, 8372.02, 8869.84, 9397.27, 9956.06, 10548.08, 11175.3, 11839.82, 12543.85, 13289.75)

	fmt.Println(notes)
	fmt.Println(len(notes))

	var file []byte
	if file, err = ioutil.ReadFile("C:\\Users\\Perrin\\Music\\soundboard\\sadviolin.mp3"); err != nil {
		//if file, err = ioutil.ReadFile("C:\\Users\\Perrin\\Music\\soundboard\\testtone.mp3"); err != nil {
		log.Fatal(err)
	}

	var dec *minimp3.Decoder
	var data []byte
	if dec, data, err = minimp3.DecodeFull(file); err != nil {
		log.Fatal(err)
	}
	fmt.Println(dec.SampleRate, dec.Channels, dec.Layer, dec.Kbps)
	fmt.Println(len(data))

	var dataleft []int16
	dataleft = make([]int16, 0)
	var dataright []int16
	dataright = make([]int16, 0)
	var lastj byte
	lastj = 0
	for i, j := range data {

		if (i % 4) == 1 {
			dataleft = append(dataleft, int16(int16(lastj)|int16(j)<<8))
		}
		if (i % 4) == 3 {
			dataright = append(dataright, int16(int16(lastj)|int16(j)<<8))
		}
		lastj = j
	}
	fmt.Println("datadone")

	/////////////////////////////////////////////////////////////////////

	fullleft := dotransform(dataleft, notes, dec.SampleRate)
	//fullright := dotransform(dataright, notes, dec.SampleRate)
	var maxes []int16
	for _, val := range fullleft {
		maxes = append(maxes, slices.Max(val))
	}
	fmt.Println(maxes)
	//fmt.Println(fullright)

	/////////////////////////////////////////////////////////////////////

	pts := make(plotter.XYs, len(dataleft))
	pts2 := make(plotter.XYs, len(dataright))

	for i, j := range dataleft {
		pts[i].X = (float64(i) / float64(dec.SampleRate))
		pts[i].Y = float64(j)
	}
	for i, j := range dataright {
		pts2[i].X = (float64(i) / float64(dec.SampleRate))
		pts2[i].Y = float64(j)
	}

	fmt.Println("plotpointsdone")
	p := plot.New()

	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err = plotutil.AddLines(p, "Left", pts, "right", pts2)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(200*vg.Inch, 4*vg.Inch, "points.svg"); err != nil {
		panic(err)
	}
}
