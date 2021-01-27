package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"runtime"

	"./topk" // borrowed a standard priority queue implementation for go
)

type QueueElement struct {
	Image image.Image
	URL   string
}

func main() {
	fmt.Println("Get top 3 colors for the following images in top3ColorsPerURL.txt")

	// input file
	input, err := os.Open("inputShort.txt")
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

	// loop through images
	queue := make(chan QueueElement, 10)
	lines := make(chan string)
	go func() {
		// NOTE: queue saturates quickly, bottlenecked by consumers,
		// so adding parallelism at producers is not beneficial in current design,
		// or until resources are added to consumers or improving consumer performance.
		for scanner.Scan() {
			url := scanner.Text()
			fmt.Println(url)
			img := getImageFromURL(url)
			if img != nil {
				queue <- QueueElement{Image: img, URL: url}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		close(queue)
	}()

	// get top 3 colors
	// run multiple goroutines to pull off the queue and process images
	// GOMAXPROCS defaults to number of cores
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for elem := range queue {
				// count colors
				colorCount := countColors(elem.Image)
				top3Colors := getTop3Colors(colorCount)
				line := fmt.Sprintf("%v, %v, %v, %v\n", elem.URL, top3Colors[0], top3Colors[1], top3Colors[2])
				lines <- line
			}
			close(lines)
		}()
	}

	// append results (url,color,color,color) to file
	for line := range lines {
		// write count to file
		if _, err = f.WriteString(line); err != nil {
			log.Fatal(err)
		}
	}
}

// getImageFromURL fetches an image and decodes it, then returns an image.Image.
func getImageFromURL(url string) image.Image {
	// fetch image
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	// decode image
	img, err := jpeg.Decode(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("decode error: ", err, " skipping ", url)
		return nil
	}
	return img
}

// countColors traverses all the pixels and creates a count of colors
// Bottlenecked here. More optimal methods could be used here such as probabilistic filtering
// (e.g. count min sketch algorithm) with tradeoff of accuracy for throughput/speed.
// Time complexity is O(N) where N is the number of pixels of image.
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

// getTop3Colors takes in the count of all colors for an image, and returns the top 3
// use a min-heap/priority queue of size 3 to keep track of top 3 colors by count
// function runs in O(N)*log(3) or O(N), where N = number of colors and 3 is the size of the minHeap
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
