package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/reujab/wallpaper"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
)

type configValues struct {
	Featured  bool
	Query     string

	//Only valid values are "landscape", "portrait", and "squarish"
	Orientation string
}

type unsplashJSON []struct {
	Urls struct {
		Full string `json:"full"`
	} `json:"urls"`

	// am going to use the download location for future versions
	Links struct {
		DownloadLocation string `json:"download_location"`
	}
}

func main() {
	directoryPath := `D:\pelle\Documents\Go workspace\go\src\github.com\pelleknaap\background changer unsplash\`

	viper.SetConfigType("yaml")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")
	viper.AddConfigPath(directoryPath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		fmt.Printf("Fatal error config file: \n%s", err)
		log.Fatal(err)
		return
	}

	config := configValues{}

	// parsing the JSON config into the struct
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("An fatal has occurred while parsing the config into a struct: \n%s", err)
		log.Fatal(err)
		return
	}

	// Getting the access key for unsplash as environment variable
	accessKey, exists := os.LookupEnv("ACCESS_KEY_UNSPLASH")
	if exists != true {
		fmt.Println("The environment variable 'ACCESS_KEY_UNSPLASH' isn't found, please make sure it's set")
		return
	}

	// Making the URL were we're going to send a request to
	//url := "https://api.unsplash.com/photos/random?count=1&featured=true&query=nature&orientation=landscape"
	url := fmt.Sprintf("https://api.unsplash.com/photos/random?count=1&featured=%v&query=%s&orientation=%s", config.Featured, config.Query, config.Orientation)

	// make a GET request to get a random picture
	jsonData, err := makeRequestToGetRandomPhoto(url, accessKey)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		return
	}

	downloadUrl := jsonData[0].Links.DownloadLocation

	// get the download link and download the image
	err = downloadFile(fmt.Sprintf("%sbackground", directoryPath), downloadUrl, accessKey)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		return
	}

	// set downloaded image as background
	err = wallpaper.SetFromFile(fmt.Sprintf("%sbackground", directoryPath))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func makeRequestToGetRandomPhoto(url string, accessKey string) (unsplashJSON, error) {
	client := &http.Client{}

	// creating the request to unsplash
	req, _ := http.NewRequest("GET", url, nil)

	//  adding the required header to authorize ourselves by unsplash
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", accessKey))

	// performing the request to unsplash
	res, err := client.Do(req)
	if err != nil {
		return unsplashJSON{}, errors.Wrap(err, "Error occurred while making request")
	}

	// making sure the body gets closed when the program exits
	defer res.Body.Close()

	var jsonData unsplashJSON

	// converting the Unsplash JSON to the image FULL URL
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&jsonData)
	if err != nil {
		return unsplashJSON{}, errors.Wrap(err, "Error occurred while decoding JSON")
	}

	return jsonData, nil
}

func downloadFile(filepath string, url string, accessKey string) error {
	client := &http.Client{}

	// creating the request to unsplash
	req, _ := http.NewRequest("GET", url, nil)

	//  adding the required header to authorize ourselves by unsplash
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", accessKey))

	// Get the image file
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "An error has occurred while getting the image file")
	}

	defer res.Body.Close()

	jsonData := struct {
		Url string
	}{}

	// converting the Unsplash JSON to the image FULL URL
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&jsonData)
	if err != nil {
		return errors.Wrap(err, "Error occurred while decoding JSON")
	}

	fmt.Printf("URL File: %s", jsonData.Url)

	// creating the request to unsplash
	fileReq, _ := http.NewRequest("GET", jsonData.Url, nil)

	// Get the image file
	resFile, err := client.Do(fileReq)
	if err != nil {
		return errors.Wrap(err, "An error has occurred while getting the image file")
	}

	defer resFile.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return errors.Wrap(err, "An error has occurred while creating the file")
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resFile.Body)
	if err != nil {
		return errors.Wrap(err, "An error has occurred while copying image file to file on disk")
	}

	return nil
}