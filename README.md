# Prompt:
----------------------------------------------------------------------------------------
Bellow is a list of links leading to an image, read this list of images and find 3 most
prevalent colors in the RGB scheme in hexadecimal format (#000000 - #FFFFFF) in each image,
and write the result into a CSV file in a form of url,color,color,color.

Please focus on speed and resources. The solution should be able to handle input files with more than a billion URLs,
using limited resources (e.g. 1 CPU, 512MB RAM). Keep in mind that there is no limit on the execution time,
but make sure you are utilizing the provided resources as much as possible at any time during the program execution.

Answer should be posted in a git repo.


# Henry's Notes:
----------------------------------------------------------------------------------------

## Initial Thoughts:

need to handle a billion URLs 
    -> must throttle to prevent resource contention (limited memory, file descriptors, network bottlenecks, etc...)
Focus
    focus on speed with limited resources
    Zaneta: do not focus on algorithms, but rather how to deal with problem at scale
CPU: can be multicore or hyperthreaded.


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


## Checklist (Personal notes):
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


## TODO:
- move hex calc out?
- close channel

## Followup Questions:
- Can we decode the image faster, or with lower quality (reduced accuracy)?
- are we allowed more probabilistic approaches?
