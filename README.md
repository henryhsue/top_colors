# How to Run:
1. `go run main.go`
This runs with shortened (30 URL) inputShort.txt. You can change to 1000 URLS swapping in intput.txt in main.go
2. after completion, view results in top3ColorsPerURL.txt (csv with url,color,color,color).
OR while running: `tail -f top3ColorsPerURL.txt`
Results are always rewritten with each invocation of `main`.


# Prompt:
Bellow is a list of links leading to an image, read this list of images and find 3 most
prevalent colors in the RGB scheme in hexadecimal format (#000000 - #FFFFFF) in each image,
and write the result into a CSV file in a form of url,color,color,color.

Please focus on speed and resources. The solution should be able to handle input files with more than a billion URLs,
using limited resources (e.g. 1 CPU, 512MB RAM). Keep in mind that there is no limit on the execution time,
but make sure you are utilizing the provided resources as much as possible at any time during the program execution.

Answer should be posted in a git repo.

Zaneta: " Please note that you'll be assessed on how you would deal with a problem at scale rather than on your ability to select the best algorithm, so focus on the problem, not the algorithms." 

# Henry's Notes:

## Initial Thoughts and Assumptions:

Need to handle a billion URLs. Must throttle to prevent resource depletion (limited memory, file descriptors, network bottlenecks, etc...)

Focus: on speed with limited resources (1CPU/512MB RAM). However, CPU can be multicore or hyperthreaded.

If problem allowed for *multiple* 1CPU/512MB RAM machines, then we can break apart problem into separate distributed 
components with multiple worker instances. e.g I do a producer -> consumer pipeline, but the producers and consumers and queue
could all be separate applications, each with 1CPU/512MB RAM. I assume the problem is not asking for this type of solution.

I also make assumptions on the size of the image. If an image is very large, it may not fit into 512MB memory, and we
will need to break apart the image into multiple subproblems. A system with a map-reduce paradigm may work better here.

## General Idea:

create a Producer - Consumer pipeline

### Producers:
download images into a queue of size X images
    perform downloading with a set of goroutines
    pause downloading when queue is full
    continue downloading images if queue drops below X

### Consumers:
pull off the queue (frees up downloading again if queue was full)
count color frequency image by image
count the most prevalent colors for an image
append to CSV file


## Checklist (Personal notes - can ignore):
- how do i get the hex value
- how do we get top k?
	count all the colors and save into a hash counter
	use a heap of size 3 to get the top 3
	traverse heap and return
- do we need to decode the image - yes.
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
- we open too many goroutines
	- due to scanner.Scan
- close channels


## Followup Questions:
- Can we decode the image faster, or with lower quality (reduced accuracy)?
- are we allowed more probabilistic approaches?
