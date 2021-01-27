/*
Bellow is a list of links leading to an image, read this list of images and find 3 most
prevalent colors in the RGB scheme in hexadecimal format (#000000 - #FFFFFF) in each image,
and write the result into a CSV file in a form of url,color,color,color.

Please focus on speed and resources. The solution should be able to handle input files with more than a billion URLs,
using limited resources (e.g. 1 CPU, 512MB RAM). Keep in mind that there is no limit on the execution time,
but make sure you are utilizing the provided resources as much as possible at any time during the program execution.

Answer should be posted in a git repo.
*/

/*
need to handle a billion URLs
focus on speed with limited resources
Zaneta: do not focus on algorithms, but rather how to deal with problem at scale
	NOTE: a bit confused by this if we're limited to 1 CPU/512MB RAM (so we're not scaling?)
download images into a queue of size 100 images
continue downloading images if queue drops below 100
process queue image by image
	count the most prevalent colors for an image
	append to CSV file

STEPS:
	download file
	open file
	read image pixel by pixel
	gather counts of colors
	sort for highest 3 counts
	get those 3 counts into a file

DONE:
- how do i get the hex value
- how do we get top k?
	count all the colors and save into a hash counter
	use a heap of size 3 to get the top 3
	traverse heap and return
- do we need to decode the image



- can we do this probabilistically?
*/

package main

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"

	"./topk"
)

func main() {
	fmt.Println("Get top 3 colors")

	// input file
	input, err := os.Open("input.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	scanner := bufio.NewScanner(input)

	// output file
	err = os.Remove("top3ColorsPerURL.txt")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile("top3ColorsPerURL.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// create queue
	/*
			TODO:
			- how many goroutines to run?
				limit by "Buffered Channel Semaphore"
			- implement priority queue
				- have goroutine download images on the side
				- place images into a queue
				- have another goroutine pull images off the queue
			- use channel for buffered queue
				- close channel after completed?
					- or let it drain
				- buffered channel
				- how to make multiple goroutines read off queue?
					- pg. 233
		    - current issues
				- we open too many goroutines
					- due to scanner.Scan
	*/
	// loop through images
	queue := make(chan image.Image, 10)
	go func() {
		for scanner.Scan() {
			url := scanner.Text()
			fmt.Println(url)

			// fetch image
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}

			// decode image
			img, err := jpeg.Decode(resp.Body)
			if err != nil {
				log.Println("decode error: ", err, " skipping ", url)
				continue
			}
			queue <- img
			fmt.Println("queue size", cap(queue))
			resp.Body.Close()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		for {
			// pop off queue
			img := <-queue

			// count colors
			colorCount := countColors(img)
			top3Colors := getTop3Colors(colorCount)
			line := fmt.Sprintf("%v, %v, %v, %v\n", "<url>", top3Colors[0], top3Colors[1], top3Colors[2])

			// write count to file
			if _, err = f.WriteString(line); err != nil {
				log.Fatal(err)
			}

		}
	}()
}

func countColors(img image.Image) map[string]int {
	colors := map[string]int{}
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			color := img.At(x, y)
			r, g, b, _ := color.RGBA()
			hex := fmt.Sprintf("#%02x%02x%02x", r, g, b)
			_, ok := colors[hex]
			if ok {
				colors[hex] += 1
			} else {
				colors[hex] = 1
			}
		}
	}
	return colors
}

func getTop3Colors(colorCount map[string]int) [3]string {
	pq := make(topk.PriorityQueue, len(colorCount))
	i := 0
	for color, count := range colorCount {
		pq[i] = &topk.Item{
			Value:    color,
			Priority: count,
			Index:    i,
		}
		i++
	}
	heap.Init(&pq)

	var top [3]string
	for i := 0; i < 3; i++ {
		if pq.Len() == 0 {
			break
		}
		item := heap.Pop(&pq).(*topk.Item)
		top[i] = item.Value
	}
	return top
}

var errInvalidFormat = errors.New("invalid format")
