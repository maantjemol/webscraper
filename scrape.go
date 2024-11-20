package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Define a struct to hold the URL and the IsAI flag
type CheckedURL struct {
	URL      string
	IsAI     bool
	keywords []string
}

func fetchURL(url string, wg *sync.WaitGroup, sem chan struct{}, checkedURLs *[]CheckedURL) {
	// Defer to signal that the goroutine has finished
	defer wg.Done()

	// Acquire a spot in the semaphore channel (limit concurrency)
	sem <- struct{}{}

	// Perform the HTTP GET request
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Error fetching URL:", err)
		<-sem // Release the spot in the semaphore
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		<-sem // Release the spot in the semaphore
		return
	}

	// Check if the body contains AI-related content (for example)
	// This is a placeholder; you can add your logic to determine if it's an AI-related URL
	isAI, keywords := containsAI(body)

	fmt.Printf("URL: %s, IsAI: %v, Keywords: %v\n", url, isAI, keywords)

	// Add the URL and its IsAI flag to the checkedURLs slice
	*checkedURLs = append(*checkedURLs, CheckedURL{URL: url, IsAI: isAI, keywords: keywords})

	// Release the spot in the semaphore
	<-sem
}

// Example function to check if the response body contains AI-related content
func containsAI(body []byte) (bool, []string) {
	// Convert the body to a string for easier text processing
	bodyStr := string(body)

	// Define a list of keywords related to AI and machine learning
	keywords := []string{
		" ai ", " ai\n", "artificial intelligence", "machine learning", "deep learning", "neural network",
		"computer vision", "natural language processing", "nlp", "reinforcement learning", "robotics",
		"chatbot", "automation", "algorithm", "predictive analytics", "big data", "cognitive computing",
		"data science", "supervised learning", "unsupervised learning",
	}

	var containsKeywords []string

	// Iterate over the keywords and check if any of them appear in the body
	for _, keyword := range keywords {
		// Using a case-insensitive search for each keyword
		if strings.Contains(strings.ToLower(bodyStr), strings.ToLower(keyword)) {
			containsKeywords = append(containsKeywords, keyword)
		}
	}

	if len(containsKeywords) > 0 {
		// If any keyword is found, return true
		return true, containsKeywords
	}

	// If no AI-related keywords are found, return false
	return false, nil
}

func main() {
	// Open SaaS.txt text file to load the URLs, separated by \n
	file, err := os.Open("SaaS.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Create a slice to store the URLs
	var urls []string

	// Loop over each line in the file
	for scanner.Scan() {
		urls = append(urls, "https://"+scanner.Text())
	}

	// Semaphore channel to limit concurrency to 50
	sem := make(chan struct{}, 50) // Buffer size is 50

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a slice to store CheckedURL structs
	var checkedURLs []CheckedURL

	// Loop over each URL and start a goroutine
	for _, url := range urls {
		wg.Add(1) // Increment the counter for the WaitGroup
		go fetchURL(url, &wg, sem, &checkedURLs)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Print all the checked URLs along with their IsAI flag
	for _, checkedURL := range checkedURLs {
		fmt.Printf("URL: %s, IsAI: %v\n", checkedURL.URL, checkedURL.IsAI)
	}

	// write the checkedURLs to a csv file
	file, err = os.Create("checkedURLs.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	for _, checkedURL := range checkedURLs {
		if checkedURL.IsAI {
			w.Write([]string{checkedURL.URL, "true", strings.Join(checkedURL.keywords, ";")})
		} else {
			w.Write([]string{checkedURL.URL, "false"})
		}
	}

	fmt.Println("All URLs have been fetched.")
}
